package usecase

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// ErrExit señala que el usuario pidió terminar la sesión mediante un
// comando slash (/exit o /quit). No es un error real: es una señal de
// control para que SessionService finalice el loop limpiamente.
var ErrExit = errors.New("sesión terminada por comando")

// CommandDispatcher interpreta y ejecuta comandos slash sobre la sesión.
//
// Desde Fase 3, ya no mantiene activeModel como estado propio (ver
// phase_two.md §8): opera sobre la *domain.Session activa, cuya
// instancia es dueña de SessionService y se comparte por puntero.
type CommandDispatcher struct {
	io      ports.SessionIO
	models  ports.ModelProvider
	runner  ports.ModelRunner
	history ports.HistoryStore
	session *domain.Session
}

func NewCommandDispatcher(io ports.SessionIO, models ports.ModelProvider, runner ports.ModelRunner, history ports.HistoryStore, session *domain.Session) *CommandDispatcher {
	return &CommandDispatcher{io: io, models: models, runner: runner, history: history, session: session}
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
	case "model":
		d.model(cmd.Args)
	case "history":
		d.historyCmd()
	case "":
		d.io.WriteLine("Comando vacío. Usa /help para ver los comandos disponibles.")
	default:
		d.io.WriteLine(fmt.Sprintf("Comando desconocido: /%s. Usa /help para ver los comandos disponibles.", cmd.Name))
	}
	return nil
}

func (d *CommandDispatcher) help() {
	d.io.WriteLine("Comandos disponibles:")
	d.io.WriteLine("  /help           - muestra esta ayuda")
	d.io.WriteLine("  /clear          - limpia el historial de mensajes de la sesión activa")
	d.io.WriteLine("  /models	      - lista los modelos de IA detectados localmente y cuáles están cargados en memoria")
	d.io.WriteLine("  /model <nombre> - selecciona el modelo activo para la sesión (ver /models)")
	d.io.WriteLine("  /history        - lista y retoma conversaciones guardadas")
	d.io.WriteLine("  /exit, /quit    - termina la sesión")
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

func (d *CommandDispatcher) model(args []string) {
	if len(args) == 0 {
		if d.session.ActiveModel == "" {
			d.io.WriteLine("No hay modelo activo. Usa /model <nombre> (ver /models list).")
			return
		}
		d.io.WriteLine("Modelo activo: " + d.session.ActiveModel)
		return
	}

	name := args[0]

	models, err := d.models.ListModels()
	if err != nil {
		d.io.WriteLine("No se pudo verificar el modelo: " + err.Error())
		return
	}

	found := false
	for _, m := range models {
		if m.Name == name {
			found = true
			break
		}
	}
	if !found {
		d.io.WriteLine(fmt.Sprintf("Modelo '%s' no está entre los detectados. Usa /models list para ver los disponibles.", name))
		return
	}

	d.session.ActiveModel = name
	d.io.WriteLine("Modelo activo cambiado a: " + name)
}

// HandlePrompt implementa RF-04: envía texto libre al modelo activo de
// la sesión y registra ambos lados del turno en Session.Messages
// (RF-08), para que quede disponible al persistir con /history.
func (d *CommandDispatcher) HandlePrompt(text string) {
	if d.session.ActiveModel == "" {
		d.io.WriteLine("No hay modelo activo. Usa /model <nombre> (ver /models list) antes de escribir.")
		return
	}

	d.session.Messages = append(d.session.Messages, domain.Message{
		Role: "user", Content: text, At: time.Now(),
	})

	d.io.WriteLine("Pensando...")
	response, err := d.runner.Generate(d.session.ActiveModel, text)
	if err != nil {
		d.io.WriteLine("Error al consultar el modelo: " + err.Error())
		return
	}

	d.session.Messages = append(d.session.Messages, domain.Message{
		Role: "assistant", Content: response, At: time.Now(),
	})
	d.io.WriteLine(response)
}

// historyCmd implementa RF-08 (comando /history): lista las
// conversaciones guardadas por FileStore, deja elegir una mediante
// selección interactiva (ports.SessionIO.SelectFromList) y la carga
// como sesión activa (mensajes + modelo usado en su momento).
func (d *CommandDispatcher) historyCmd() {
	metas, err := d.history.List()
	if err != nil {
		d.io.WriteLine("No se pudo leer el historial de conversaciones: " + err.Error())
		return
	}

	if len(metas) == 0 {
		d.io.WriteLine("No hay conversaciones guardadas todavía.")
		return
	}

	options := make([]domain.ListOption, len(metas))
	for i, m := range metas {
		title := m.Title
		if title == "" {
			title = "(sin título)"
		}
		summary := m.Summary
		if summary == "" {
			summary = "sin resumen"
		}
		options[i] = domain.ListOption{
			Title:       title,
			Description: fmt.Sprintf("%s — %s", m.CreatedAt.Format("02/01/2006 15:04"), summary),
		}
	}

	idx, err := d.io.SelectFromList("Conversaciones guardadas:", options)
	if err != nil {
		if errors.Is(err, ports.ErrSelectionCancelled) {
			d.io.WriteLine("Selección cancelada.")
			return
		}
		d.io.WriteLine("No se pudo completar la selección: " + err.Error())
		return
	}

	loaded, err := d.history.Load(metas[idx].ID)
	if err != nil {
		d.io.WriteLine("No se pudo cargar la conversación: " + err.Error())
		return
	}

	*d.session = loaded

	activeModel := d.session.ActiveModel
	if activeModel == "" {
		activeModel = "ninguno"
	}
	d.io.WriteLine(fmt.Sprintf("Conversación '%s' cargada (%d mensajes, modelo: %s).",
		d.session.Title, len(d.session.Messages), activeModel))

	d.replayTranscript()
}

// replayTranscript reimprime los mensajes de la sesión recién cargada en
// el historial visual de la terminal (RF-08: "retomar/cargar una
// conversación"). d.session.Messages ya queda poblado con solo asignar
// *d.session = loaded — eso alcanza para que el modelo tenga contexto en
// el próximo prompt, pero el historial visual (viewport) es un []string
// aparte que vive en terminal/model.go (ver intermediate_phases.md,
// notas de Fase 2.5) y no se sincroniza solo. Esta función lo hace
// pasando cada mensaje por el mismo ports.SessionIO.WriteLine que ya se
// usa en todos lados, sin que command_dispatcher.go conozca nada de
// bubbletea/viewport (la frontera hexagonal no se toca).
func (d *CommandDispatcher) replayTranscript() {
	if len(d.session.Messages) == 0 {
		return
	}

	d.io.WriteLine("")
	d.io.WriteLine("── Transcripción de la conversación cargada ──")
	for _, m := range d.session.Messages {
		switch m.Role {
		case "user":
			d.io.WriteLine("")
			d.io.WriteLine("Tú: " + m.Content)
		case "assistant":
			d.io.WriteLine("")
			d.io.WriteLine("Argos: " + m.Content)
		}
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

// clear implementa RF-08 (ya no es stub, resuelve el TODO(fase 3) que
// venía desde Fase 1/2): limpia los mensajes de la sesión activa en
// memoria. No toca lo ya persistido en .argos/sessions/ ni el modelo
// activo — solo el historial de la conversaciFón en curso.
func (d *CommandDispatcher) clear() {
	d.session.Messages = nil
	d.io.WriteLine("Historial de la conversación limpiado.")
}
