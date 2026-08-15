package main

import (
	"fmt"
	"os"

	"github.com/pokydev/argos/internal/adapters/ollama"
	"github.com/pokydev/argos/internal/adapters/scanner"
	"github.com/pokydev/argos/internal/adapters/storage"
	"github.com/pokydev/argos/internal/adapters/terminal"
	"github.com/pokydev/argos/internal/core/usecase"
)

func main() {
	term := terminal.New()
	defer term.Close()

	models := ollama.New()
	projectScanner := scanner.New()

	// RF-08/RF-09: la raíz del historial y de /init es el directorio de
	// trabajo actual. Resolver --path (RF-12) queda pendiente para un
	// checklist futuro.
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

	session := usecase.NewSessionService(term, models, models, history, projectScanner, root)

	if err := session.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
