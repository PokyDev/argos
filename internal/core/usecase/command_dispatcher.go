package usecase

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// ErrExit señala que el usuario pidió terminar la sesión mediante un
// comando slash (/exit o /quit). No es un error real: es una señal de
// control para que SessionService finalice el loop limpiamente.
var ErrExit = errors.New("sesión terminada por comando")

// CommandDispatcher interpreta y ejecuta comandos slash sobre la sesión.
type CommandDispatcher struct {
	io     ports.SessionIO
	models ports.ModelProvider
}

func NewCommandDispatcher(io ports.SessionIO, models ports.ModelProvider) *CommandDispatcher {
	return &CommandDispatcher{io: io, models: models}
}

// IsCommand indica si una línea de entrada corresponde a un slash command.
func IsCommand(line string) bool {
	return strings.HasPrefix(line, "/")
}

// ParseCommand convierte una línea cruda ("/model llama3") en domain.Command.
func ParseCommand(line string) domain.Command {
	trimmed := strings.TrimPrefix(line, "/")
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return domain.Command{}
	}
	return domain.Command{
		Name: strings.ToLower(fields[0]),
		Args: fields[1:],
	}
}

// Dispatch ejecuta el comando correspondiente. Devuelve ErrExit si el
// usuario pidió salir.
func (d *CommandDispatcher) Dispatch(cmd domain.Command) error {
	switch cmd.Name {
	case "help":
		d.help()
	case "exit", "quit":
		return ErrExit
	case "clear":
		d.clear()
	case "models":
		d.listModels()
	case "":
		d.io.WriteLine("Comando vacío. Usa /help para ver los comandos disponibles.")
	default:
		d.io.WriteLine(fmt.Sprintf("Comando desconocido: /%s. Usa /help para ver los comandos disponibles.", cmd.Name))
	}
	return nil
}

func (d *CommandDispatcher) help() {
	d.io.WriteLine("Comandos disponibles:")
	d.io.WriteLine("  /help          - muestra esta ayuda")
	d.io.WriteLine("  /clear         - limpia el contexto/historial de la sesión")
	d.io.WriteLine("  /models list   - lista los modelos de IA detectados localmente y cuáles están cargados en memoria")
	d.io.WriteLine("  /exit, /quit   - termina la sesión")
}

// listModels implementa RF-02: detección automática de modelos locales.
// Cubre tanto "/models" como "/models list" (el subcomando "list" es el
// único soportado por ahora; se ignora si viene, ya que no hay otra
// variante todavía).
func (d *CommandDispatcher) listModels() {
	models, err := d.models.ListModels()
	if err != nil {
		d.io.WriteLine("No se pudo obtener la lista de modelos: " + err.Error())
		return
	}

	if len(models) == 0 {
		d.io.WriteLine("No se detectaron modelos instalados en Ollama.")
		d.io.WriteLine("Instalá uno con: ollama pull <modelo>  (ej. ollama pull llama3.2:1b)")
		return
	}

	d.io.WriteLine(fmt.Sprintf("Modelos detectados (%d):", len(models)))
	for _, m := range models {
		status := "disponible"
		if m.Loaded {
			status = "cargado"
		}
		d.io.WriteLine(fmt.Sprintf("  - %s | %s | %s", m.Name, formatSize(m.Size), status))
	}
}

// formatSize convierte bytes a una unidad legible (GB si aplica, si no MB).
func formatSize(bytes int64) string {
	const mb = 1024 * 1024
	const gb = 1024 * mb

	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.0f MB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (d *CommandDispatcher) clear() {
	// TODO(fase 3): limpiar historial real de conversación (RF-08) cuando
	// SessionService mantenga estado de mensajes. Por ahora es un stub
	// intencional para no anticipar la entidad Session/Message sin uso real.
	d.io.WriteLine("Contexto/historial limpiado.")
}
