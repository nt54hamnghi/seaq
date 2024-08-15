package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	textinput.Model
	history
}

func New() Model {
	return Model{
		Model:   textinput.New(),
		history: newHistory(),
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.SetValue(m.previous())
			m.CursorEnd()
			return m, nil
		case tea.KeyDown:
			m.SetValue(m.next())
			m.CursorEnd()
			return m, nil
		}
	}

	return m, cmd
}

func (m Model) View() string {
	return m.Model.View()
}

func (m *Model) Append(value string) {
	m.history.append(value)
	m.Reset()
}
