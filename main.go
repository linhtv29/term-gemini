package main

import (
	"log"
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx := context.Background()
	gemini := NewGemini(ctx)
	defer gemini.client.Close()
	p := tea.NewProgram(InitialModel(gemini))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)
