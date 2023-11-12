package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nbd-wtf/go-nostr"
)

type model struct {
	ctx         context.Context
	relayUrl    string
	groupId     string
	viewport    viewport.Model
	messages    messagesbox
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
}

type messagesbox struct {
	events []nostr.Event
	total  int
}

func (ms messagesbox) render() string {
	sb := strings.Builder{}
	sb.Grow(ms.total)
	for _, evt := range ms.events {
		sb.WriteString(evt.Content + "\n")
	}
	return sb.String()
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "type a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`loading chat messages`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		ctx:         context.Background(),
		relayUrl:    "wss://dev2.hazlitt.fiatjaf.com",
		groupId:     "z",
		textarea:    ta,
		messages:    messagesbox{events: make([]nostr.Event, 0, 120)},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			go sendChatMessage(m, m.textarea.Value())
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case messagesbox:
		m.messages = msg
		m.viewport.SetContent(m.messages.render())
	case *nostr.Event:
		m.messages.events = append(m.messages.events, *msg)
		m.messages.total += len(msg.Content)
		m.viewport.SetContent(m.messages.render())
	case error:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
