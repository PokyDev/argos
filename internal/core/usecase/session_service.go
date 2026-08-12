package usecase

import (
	"errors"
	"io"

	"github.com/pokydev/argos/internal/core/ports"
)

type SessionService struct {
	io         ports.SessionIO
	dispatcher *CommandDispatcher
}

func NewSessionService(io ports.SessionIO, models ports.ModelProvider) *SessionService {
	return &SessionService{
		io:         io,
		dispatcher: NewCommandDispatcher(io, models),
	}
}

func (s *SessionService) Run() error {
	s.io.WriteLine("Argos CLI iniciada. Escribe /help para ver los comandos disponibles.")

	for {
		line, err := s.io.ReadLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
					s.io.WriteLine("Hasta luego.")
					return nil
				}
				return err
			}
			continue
		}

		// Placeholder: la conexión con el modelo local (Ollama) llega en
		// la Fase 2. Por ahora, texto libre solo hace eco.
		s.io.WriteLine("(echo) " + line)
	}
}
