package input

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	promptIcon        = "> "
	promptColor       = lipgloss.Color("#66b3ff")
	promptPlaceHolder = "What's in your mind?"
	defaultCharLimit  = 4096
)

var promptStyle = lipgloss.NewStyle().Foreground(promptColor)

type Model struct {
	textinput.Model
	history
}

func New() *Model {
	// create & configure model
	model := textinput.New()
	model.Focus()                                 // Immediately focus the input
	model.Placeholder = promptPlaceHolder         // Set placeholder text
	model.Prompt = promptStyle.Render(promptIcon) // Set styled prompt
	model.CharLimit = defaultCharLimit

	return &Model{
		Model:   model,
		history: newHistory(),
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)

	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.Type {
		case tea.KeyUp:
			m.SetValue(m.previous())
			m.CursorEnd()
			return m, nil
		case tea.KeyDown:
			m.SetValue(m.next())
			m.CursorEnd()
			return m, nil
		default:
			return m, cmd
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if !m.Focused() {
		return ""
	}
	return m.Model.View()
}

func (m Model) AsString() string {
	return fmt.Sprintf("%s%s", promptStyle.Render(promptIcon), m.Value())
}

func (m *Model) Append(value string) {
	m.history.append(value)
	m.Reset()
}
