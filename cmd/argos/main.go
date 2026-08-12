package main

import (
	"fmt"
	"os"

	"github.com/pokydev/argos/internal/adapters/ollama"
	"github.com/pokydev/argos/internal/adapters/terminal"
	"github.com/pokydev/argos/internal/core/usecase"
)

func main() {
	term := terminal.New()
	defer term.Close()

	models := ollama.New()

	session := usecase.NewSessionService(term, models)

	if err := session.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
