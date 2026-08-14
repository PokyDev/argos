package usecase

import (
	"errors"
	"io"

	"github.com/pokydev/argos/internal/core/domain"
	"github.com/pokydev/argos/internal/core/ports"
)

// SessionService orquesta el loop interactivo (REPL) y es, desde Fase
// 3, la dueña de la *domain.Session activa (RF-08): la crea al arrancar
// y la comparte por puntero con CommandDispatcher, que la lee/modifica
// pero no es dueño de su ciclo de vida.
type SessionService struct {
	io      ports.SessionIO
	history ports.HistoryStore
	session *domain.Session

	dispatcher *CommandDispatcher
}

func NewSessionService(io ports.SessionIO, models ports.ModelProvider, runner ports.ModelRunner, history ports.HistoryStore) *SessionService {
	session := domain.NewSession()
	return &SessionService{
		io:         io,
		history:    history,
		session:    session,
		dispatcher: NewCommandDispatcher(io, models, runner, history, session),
	}
}

func (s *SessionService) Run() error {
	s.io.WriteLine("Argos CLI iniciada. Escribe /help para ver los comandos disponibles.")

	for {
		line, err := s.io.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.persistSession()
				s.io.WriteLine("Sesión finalizada (EOF).")
				return nil
			}
			return err
		}

		if line == "" {
			continue
		}

		if IsCommand(line) {
			cmd := ParseCommand(line)
			if err := s.dispatcher.Dispatch(cmd); err != nil {
				if errors.Is(err, ErrExit) {
					s.persistSession()
					s.io.WriteLine("Hasta luego.")
					return nil
				}
				return err
			}
			continue
		}

		// RF-04: texto libre se envía al modelo activo de la sesión.
		s.dispatcher.HandlePrompt(line)
	}
}

// persistSession guarda la sesión activa al cerrar (RF-08), siempre y
// cuando tenga al menos un mensaje: una sesión vacía (el usuario abrió
// y cerró argos sin escribir nada) no se guarda, para no llenar
// .argos/sessions/ de archivos sin contenido útil para /history.
func (s *SessionService) persistSession() {
	if len(s.session.Messages) == 0 {
		return
	}

	summarizeSession(s.dispatcher.runner, s.session)

	if err := s.history.Save(*s.session); err != nil {
		s.io.WriteLine("No se pudo guardar el historial de la conversación: " + err.Error())
	}
}
