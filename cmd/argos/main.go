package main

import (
	"fmt"
	"os"

	"github.com/pokydev/argos/internal/adapters/ollama"
	"github.com/pokydev/argos/internal/adapters/storage"
	"github.com/pokydev/argos/internal/adapters/terminal"
	"github.com/pokydev/argos/internal/core/usecase"
)

func main() {
	term := terminal.New()
	defer term.Close()

	models := ollama.New()

	// RF-08: la raíz del historial es el directorio de trabajo actual.
	// Resolver --path (RF-12) queda para el checklist de /init.
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	history, err := storage.New(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	session := usecase.NewSessionService(term, models, models, history)

	if err := session.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
