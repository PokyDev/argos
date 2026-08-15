package usecase

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// fileSummary asocia un archivo escaneado con el resumen que el modelo
// activo generó para él. Es un detalle interno de runInit, no domain:
// no se persiste ni se usa fuera de la generación de ARGOS.md.
type fileSummary struct {
	path    string
	summary string
}

// runInit implementa la orquestación de /init (RF-09): recorre el
// proyecto activo con el scanner inyectado, le pide al modelo activo un
// resumen por archivo, sintetiza esos resúmenes en un párrafo general y
// persiste todo como ARGOS.md en la raíz del proyecto (Session.RootPath).
//
// Vive en su propio archivo (no en command_dispatcher.go) porque
// encadena varias llamadas a Generate — mismo criterio anotado en
// project_description.md §6 para justificar project_init.go.
//
// Validación previa (decisión explícita, ver nota de sesión en
// phase_three.md): a diferencia del diseño original en
// project_description.md §6.2 (que preveía selección interactiva de
// modelo dentro del propio /init, igual que /history), acá se exige que
// el modelo ya esté activo vía /model antes de correr /init. Simplifica
// el flujo y reutiliza el mismo mensaje de validación que ya usa
// HandlePrompt.
func runInit(io ports.SessionIO, ps ports.ProjectScanner, runner ports.ModelRunner, session *domain.Session) {
	if session.ActiveModel == "" {
		io.WriteLine("No hay modelo activo. Usá /model <nombre> (ver /models) antes de ejecutar /init.")
		return
	}
	if session.RootPath == "" {
		io.WriteLine("No se pudo determinar la raíz del proyecto para /init.")
		return
	}

	io.WriteLine("Escaneando archivos del proyecto en " + session.RootPath + "...")
	files, truncated, err := ps.Walk(session.RootPath)
	if err != nil {
		io.WriteLine("No se pudo escanear el proyecto: " + err.Error())
		return
	}
	if len(files) == 0 {
		io.WriteLine("No se encontraron archivos de texto para analizar (¿directorio vacío, o todo excluido/binario?).")
		return
	}

	io.WriteLine(fmt.Sprintf("Analizando %d archivo(s) con el modelo activo (%s). Esto puede tardar unos minutos con modelos chicos/CPU...", len(files), session.ActiveModel))

	summaries := make([]fileSummary, 0, len(files))
	for i, f := range files {
		io.WriteLine(fmt.Sprintf("  [%d/%d] %s", i+1, len(files), f.Path))

		raw, err := runner.Generate(session.ActiveModel, fileSummaryPrompt(f))
		if err != nil {
			// RNF-06: que un archivo falle no debe abortar todo /init.
			summaries = append(summaries, fileSummary{path: f.Path, summary: "(no se pudo analizar: " + err.Error() + ")"})
			continue
		}
		summaries = append(summaries, fileSummary{path: f.Path, summary: strings.TrimSpace(raw)})
	}

	io.WriteLine("Sintetizando resumen general del proyecto...")
	overview, err := runner.Generate(session.ActiveModel, overviewPrompt(summaries))
	if err != nil {
		io.WriteLine("No se pudo generar el resumen general (ARGOS.md se genera igual, solo con el detalle por archivo): " + err.Error())
		overview = ""
	}

	content := buildArgosMarkdown(overview, summaries, truncated)

	if err := ps.WriteContext(session.RootPath, content); err != nil {
		io.WriteLine("No se pudo guardar ARGOS.md: " + err.Error())
		return
	}

	io.WriteLine(fmt.Sprintf("ARGOS.md generado en %s.", filepath.Join(session.RootPath, "ARGOS.md")))
}

// fileSummaryPrompt arma el prompt para pedirle al modelo un resumen
// corto de un archivo individual. truncate() (definida en history.go,
// mismo paquete) acota el contenido para no desbordar el contexto de
// modelos chicos con archivos largos.
func fileSummaryPrompt(f domain.ScannedFile) string {
	var b strings.Builder
	b.WriteString("Resumí en 1 o 2 oraciones cuál es el propósito o responsabilidad de este archivo dentro de un proyecto de software. ")
	b.WriteString("Respondé solo con el resumen, sin repetir el nombre del archivo ni agregar texto extra.\n\n")
	b.WriteString("Archivo: " + f.Path + "\n\n")
	b.WriteString("Contenido:\n")
	b.WriteString(truncate(f.Content, 4000))
	return b.String()
}

// overviewPrompt arma el prompt de síntesis final a partir de los
// resúmenes ya generados por archivo (no vuelve a mandar el contenido
// crudo de cada archivo, para no repetir todo el contexto ya usado).
func overviewPrompt(summaries []fileSummary) string {
	var b strings.Builder
	b.WriteString("A partir de estos resúmenes de archivos de un proyecto de software, ")
	b.WriteString("escribí un párrafo breve (3 a 5 oraciones) explicando de qué trata el proyecto y su arquitectura general. ")
	b.WriteString("Sin encabezados, sin volver a listar los archivos uno por uno, solo el párrafo.\n\n")
	for _, fsum := range summaries {
		b.WriteString(fmt.Sprintf("- %s: %s\n", fsum.path, truncate(fsum.summary, 200)))
	}
	return b.String()
}

// buildArgosMarkdown arma el contenido final de ARGOS.md de forma
// determinística en Go (no le pide al modelo que arme el Markdown
// completo, para no depender de que un modelo chico respete formato
// complejo — mismo criterio que summarizeSession en history.go).
func buildArgosMarkdown(overview string, summaries []fileSummary, wasTruncated bool) string {
	var b strings.Builder
	b.WriteString("# ARGOS.md\n\n")
	b.WriteString("> Generado automáticamente por `/init` de Argos CLI. Se sobreescribe por completo cada vez que se vuelve a correr `/init` (sin merge incremental — ver project_description.md §6).\n\n")

	if overview != "" {
		b.WriteString("## Resumen del proyecto\n\n")
		b.WriteString(strings.TrimSpace(overview) + "\n\n")
	}

	b.WriteString("## Archivos\n\n")
	for _, fsum := range summaries {
		summary := fsum.summary
		if summary == "" {
			summary = "(sin resumen)"
		}
		b.WriteString(fmt.Sprintf("- **%s**: %s\n", fsum.path, summary))
	}

	if wasTruncated {
		b.WriteString("\n> Nota: el proyecto tiene más archivos de texto candidatos que los analizados; se aplicó un límite para mantener /init en un tiempo razonable.\n")
	}

	return b.String()
}
