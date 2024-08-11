package repl

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nt54hamnghi/hiku/pkg/repl/input"
)

type Repl struct {
	input input.Model
	chat  []ChatMsg
}

func New() Repl {
	// create & configure in
	in := input.New()
	in.Focus()
	in.Placeholder = "What's in your mind?"
	in.CharLimit = 128

	// create & configure viewport
	vp := viewport.New(0, 0)
	vp.KeyMap.Up.SetEnabled(false)
	vp.KeyMap.Down.SetEnabled(false)

	return Repl{
		input: in,
		chat:  make([]ChatMsg, 0),
	}
}

func (m Repl) Init() tea.Cmd {
	return m.input.Init()
}

func (m Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.input.Update(msg)
	if input, ok := model.(input.Model); ok {
		m.input = input
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.input.Width = (msg.Width / 3) * 2
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			switch strings.ToLower(m.input.Value()) {
			case "exit", "bye":
				return m, tea.Quit
			case "save":
				panic("todo")
			default:
				prompt := m.input.Value()
				m.input.Append(prompt)
				return m, SendChatMsg(prompt)
			}
		}
	case ChatMsg:
		m.chat = append(m.chat, msg)
	}

	return m, cmd
}

func (m Repl) View() string {
	content := JoinChatMsg(m.chat)
	return content + m.input.View() + "\n"
}
