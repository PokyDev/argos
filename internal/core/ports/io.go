package ports

import (
	"errors"

	"github.com/pokydev/argos/internal/core/domain"
)

// ErrSelectionCancelled señala que el usuario canceló una selección
// interactiva (Esc o Ctrl+C) en vez de confirmar un ítem. Mismo patrón
// que ErrExit en command_dispatcher.go: no es un error real, es una
// señal de control para que el caller (ej. /history) responda distinto
// a un fallo genuino.
var ErrSelectionCancelled = errors.New("selección cancelada por el usuario")

// SessionIO define el contrato de entrada/salida que el core necesita
// para interactuar con el usuario, sin conocer el mecanismo concreto
// (terminal hoy, voz o socket remoto en el futuro).
type SessionIO interface {
	// ReadLine bloquea hasta recibir una línea de entrada del usuario.
	// Devuelve io.EOF cuando no hay más entrada disponible (ej. Ctrl+D).
	ReadLine() (string, error)

	// WriteLine imprime una línea de salida hacia el usuario.
	WriteLine(msg string)

	// SelectFromList muestra una lista interactiva y bloquea hasta que
	// el usuario confirma un ítem (Enter) o cancela (Esc/Ctrl+C).
	// Devuelve el índice elegido dentro de options, o
	// ErrSelectionCancelled si no hubo selección. Usado hoy por
	// /history (elegir conversación) y pensado para reutilizarse en
	// /init (elegir modelo) — ver project_description.md §6.
	SelectFromList(prompt string, options []domain.ListOption) (int, error)

	// Close libera recursos del adaptador (si aplica).
	Close() error
}
