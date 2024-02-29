package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	provider    AiProvider
	prompt      string
	isLoading   bool
	spinner     spinner.Model
}

type AiProvider interface {
	SendMessage(text string) string
}

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

func InitialModel(provider AiProvider) Model {
	ta := textarea.New()
	ta.Placeholder = "Send a prompt..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(80)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(80, 3)
	vp.SetContent(`
	Hello!
	How can i help you?
	`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	modelSpinner := spinner.New()
	modelSpinner.Spinner = spinner.Points
	modelSpinner.Style = spinnerStyle

	return Model{
		textarea:    ta,
		messages:    []string{},
		prompt:      "",
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#DC9D0A")),
		err:         nil,
		provider:    provider,
		isLoading:   false,
		spinner:     modelSpinner,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	hostName, _ := os.Hostname()

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.prompt = m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render(hostName+": ")+m.prompt)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			m.textarea.Reset()
			m.isLoading = true
			return m, sendPrompt(m)
		}

	case resultMsg:
		m.messages = append(m.messages, m.senderStyle.Render("Bot: ")+string(msg))
		m.isLoading = false
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.Height = m.viewport.TotalLineCount()
		m.viewport.GotoBottom()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	spinner := ""
	if m.isLoading {
		spinner = m.spinner.View()
	}
	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		m.viewport.View(),
		spinner,
		m.textarea.View(),
	) + "\n\n"
}

func sendPrompt(m Model) tea.Cmd {
	return func() tea.Msg {
		res := m.provider.SendMessage(m.prompt)
		return resultMsg(res)
	}
}

type resultMsg string
