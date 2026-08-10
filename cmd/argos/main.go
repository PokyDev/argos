package main

import (
	"fmt"
	"os"

	"github.com/pokydev/argos/internal/adapters/terminal"
	"github.com/pokydev/argos/internal/core/usecase"
)

func main() {
	term := terminal.New()
	defer term.Close()

	session := usecase.NewSessionService(term)

	if err := session.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
