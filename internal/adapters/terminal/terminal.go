package terminal

import (
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// Terminal implementa ports.SessionIO sobre una TUI bubbletea de dos
// regiones (historial con scroll + input fijo, RF-14/RNF-12).
//
// Puente ReadLine/WriteLine <-> loop de eventos de bubbletea: ver
// comentarios en New(), ReadLine() y WriteLine().
type Terminal struct {
	prog    *tea.Program
	inputCh chan string
	doneCh  chan struct{}
}

// New arranca el programa bubbletea en background y devuelve el adaptador
// listo para usarse desde SessionService, con la misma firma que en Fase 1.
func New() *Terminal {
	inputCh := make(chan string, 16)
	doneCh := make(chan struct{})

	t := &Terminal{inputCh: inputCh, doneCh: doneCh}
	t.prog = tea.NewProgram(
		newModel(inputCh),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		defer close(doneCh)
		// Run() bloquea hasta recibir tea.Quit (Ctrl+C en la TUI, o
		// Close() desde main.go). El error de retorno solo indica
		// problemas del terminal en sí (ej. correr sin TTY real).
		_, _ = t.prog.Run()
	}()

	return t
}

func (t *Terminal) ReadLine() (string, error) {
	select {
	case line, ok := <-t.inputCh:
		if !ok {
			return "", io.EOF
		}
		return line, nil
	case <-t.doneCh:
		return "", io.EOF
	}
}

func (t *Terminal) WriteLine(msg string) {
	t.prog.Send(outputLineMsg(msg))
}

// SelectFromList implementa ports.SessionIO. Puente análogo al de
// ReadLine/WriteLine: empuja un showListMsg al tea.Program (que pasa el
// model a modo lista) y bloquea leyendo un canal propio de esta llamada
// hasta que el usuario confirma o cancela. El canal viaja dentro del
// mensaje en vez de ser un campo de Terminal porque SelectFromList no
// se espera concurrente (mismo patrón "un turno a la vez" que ReadLine).
func (t *Terminal) SelectFromList(prompt string, options []domain.ListOption) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no hay opciones para seleccionar")
	}

	resultCh := make(chan selectResult, 1)
	t.prog.Send(showListMsg{prompt: prompt, options: options, result: resultCh})

	select {
	case r := <-resultCh:
		return r.index, r.err
	case <-t.doneCh:
		return -1, ports.ErrSelectionCancelled
	}
}

func (t *Terminal) Close() error {
	t.prog.Quit()
	<-t.doneCh
	return nil
}
