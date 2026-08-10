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
	io ports.SessionIO
}

func NewCommandDispatcher(io ports.SessionIO) *CommandDispatcher {
	return &CommandDispatcher{io: io}
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
	d.io.WriteLine("  /exit, /quit   - termina la sesión")
}

func (d *CommandDispatcher) clear() {
	// TODO(fase 3): limpiar historial real de conversación (RF-08) cuando
	// SessionService mantenga estado de mensajes. Por ahora es un stub
	// intencional para no anticipar la entidad Session/Message sin uso real.
	d.io.WriteLine("Contexto/historial limpiado.")
}
