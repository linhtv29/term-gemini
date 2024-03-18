package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	// "sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

type Model struct {
	viewport       viewport.Model
	messages       []string
	textarea       textarea.Model
	senderStyle    lipgloss.Style
	err            error
	provider       AiProvider
	prompt         string
	isLoading      bool
	isShowTextarea bool
	ready          bool
	spinner        spinner.Model
}

type AiProvider interface {
	SendMessage(text string) string
}

var secondaryColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#43F55E"))
var dangerColor = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))

func InitialModel(provider AiProvider) Model {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}

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
	modelSpinner.Style = secondaryColor

	return Model{
		textarea:       ta,
		messages:       []string{},
		prompt:         "",
		senderStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#DC9D0A")),
		err:            nil,
		provider:       provider,
		ready:          false,
		isLoading:      false,
		isShowTextarea: true,
		spinner:        modelSpinner,
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
			m.showLoading()
			m.viewport.GotoBottom()
			return m, sendPrompt(m)
		}

		switch msg.String() {
		case "ctrl+s":
			prompt := "translate to Vietnamese: "
			m = m.checkTextarea(prompt)
			m.showLoading()
			return m, translate(m)
		case "ctrl+e":
			prompt := "translate to English: "
			m = m.checkTextarea(prompt)
			m.showLoading()
			return m, translate(m)
		case "ctrl+g":
			m = m.checkTextarea("")
			m.showLoading()
			return m, checkGrammar(m)
		case "i":
			m.isShowTextarea = true
			if m.textarea.Value() == "i" {
				m.textarea.Reset()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height - 6)
			// m.viewport.YPosition = 3
			m.viewport.SetContent("Type some shit!")
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height
		}

	case resultMsg:
		m.messages = append(m.messages, secondaryColor.Render("Bot: ")+string(msg))
		m.isLoading = false
		m.isShowTextarea = false
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.resizeViewport()

	case grammarResult:
		if !msg.Correct {
			m.messages = append(m.messages, dangerColor.Render("Có lỗi ngữ pháp!"))
			m.messages = append(m.messages, m.senderStyle.Render("Bản gốc: ")+string(msg.Origin)+"\n")
			m.messages = append(m.messages, secondaryColor.Render(msg.Explanation)+"\n")
			m.messages = append(m.messages, m.senderStyle.Render("Sửa lại => ")+string(msg.Fixed))
		} else {
			m.messages = append(m.messages, secondaryColor.Render("Không có lỗi!\n"))
			m.messages = append(m.messages, m.senderStyle.Render("Bản gốc: ")+string(msg.Origin))
		}
		m.isLoading = false
		m.isShowTextarea = false
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.resizeViewport()

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
	textarea := ""
	if m.isLoading {
		spinner = m.spinner.View()
	}
	if m.isShowTextarea {
		textarea = m.textarea.View()
	}
	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		m.viewport.View(),
		spinner,
		textarea,
	) + "\n\n"
}

func (m *Model) showLoading() {
	m.isLoading = true
	m.textarea.Reset()
}
func (m *Model) resizeViewport() {
	m.viewport.Height = m.viewport.TotalLineCount() + 2
	m.viewport.GotoBottom()
}

func (m Model) checkTextarea(prompt string) Model {
	clipDocs := string(clipboard.Read(clipboard.FmtText))
	if m.textarea.Value() != "" {
		m.prompt = prompt + m.textarea.Value()
	} else {
		m.prompt = prompt + clipDocs
	}
	return m
}

func sendPrompt(m Model) tea.Cmd {
	return func() tea.Msg {
		res := m.provider.SendMessage(m.prompt)
		return resultMsg(res)
	}
}

func translate(m Model) tea.Cmd {
	return func() tea.Msg {
		res := m.provider.SendMessage(m.prompt)
		clipboard.Write(clipboard.FmtText, []byte(res))
		return resultMsg(res)
	}
}

func checkGrammar(m Model) tea.Cmd {
	var checkResult grammarResult
	prompt := "Grammar check: " + m.prompt + ". Response a result in json object with syntax: {\"correct\": true/false, \"falseWords\": [{\"word\": \"falseWord1\", \"index\": index of falseWord1 in sentence separator by space}, {\"word\": \"falseWord2\", \"index\": index of falseWord2 in sentence separator by space}], \"fixed\": \"grammar fixed\" , \"explanation\": \"explanation grammar errors by vietnamese\" }"

	// USE waitgroup to wait for the goroutine

	// var wg sync.WaitGroup
	// wg.Add(1)
	// return func() tea.Msg {
	// 	res := m.provider.SendMessage(prompt)
	// 	go func() {
	// 		defer wg.Done()
	// 		err := json.Unmarshal([]byte(string(res)), &checkResult)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 	}()
	// 	wg.Wait()
	// 	checkResult.Origin = m.prompt
	// 	clipboard.Write(clipboard.FmtText, []byte(checkResult.Fixed))
	// 	return checkResult
	// }

	// USE channel to wait for the goroutine
	ch := make(chan grammarResult)
	return func() tea.Msg {
		response := m.provider.SendMessage(prompt)
		go func() {
			err := json.Unmarshal([]byte(string(response)), &checkResult)
			if err != nil {
				fmt.Println(err)
			}
			checkResult.Origin = m.prompt
			clipboard.Write(clipboard.FmtText, []byte(checkResult.Fixed))
			ch <- checkResult
		}()
		select {
		case result := <-ch:
			return result
		case <-time.After(10 * time.Second):
			return nil
		}
	}
}

type resultMsg string
type grammarResult struct {
	Correct    bool `json:"correct"`
	FalseWords []struct {
		Word  string `json:"word"`
		Index int    `json:"index"`
	} `json:"falseWords"`
	Explanation string `json:"explanation"`
	Fixed       string `json:"fixed"`
	Origin      string `json:"origin"`
}
