package main

import (
	"context"
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx := context.Background()
	gemini := NewGemini(ctx)
	defer gemini.client.Close()
	p := tea.NewProgram(InitialModel(gemini), tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)
