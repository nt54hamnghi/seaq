package repl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nt54hamnghi/hiku/pkg/rag"
	"github.com/nt54hamnghi/hiku/pkg/repl/input"
	"github.com/tmc/langchaingo/schema"
)

type ErrMsg struct {
	error
}

func (e ErrMsg) Error() string {
	return e.error.Error()
}

func (e ErrMsg) Unwrap() error {
	return e.error
}

type StoreMsg struct {
	*rag.DocumentStore
}

func InitStore(ctx context.Context, docs []schema.Document) tea.Cmd {
	return func() tea.Msg {
		if len(docs) == 0 {
			return ErrMsg{errors.New("no documents to load")}
		}

		store, err := rag.NewStoreWithDocuments(ctx, docs)
		if err != nil {
			return ErrMsg{err}
		}

		return StoreMsg{store}
	}
}

type Repl struct {
	input    input.Model        // Input UI
	viewport viewport.Model     // Viewport UI for chat history
	chat     []ChatMsg          // Chat history
	error    error              // Critical error
	store    *rag.DocumentStore // Data store
	ctx      context.Context
}

func New(ctx context.Context, docs []schema.Document) Repl {
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
		input:    in,
		viewport: vp,
		chat:     make([]ChatMsg, 0),
		ctx:      ctx,
		store:    &rag.DocumentStore{Docs: docs},
	}
}

func (m Repl) Init() tea.Cmd {
	return tea.Batch(
		m.input.Init(),
		InitStore(m.ctx, m.store.Docs),
	)
}

func (m Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		vpCmd tea.Cmd
		tiCmd tea.Cmd
	)

	in, tiCmd := m.input.Update(msg)
	if input, ok := in.(input.Model); ok {
		m.input = input
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width / 3 * 2
		m.input.Width = w
		m.viewport.Width = w
		m.viewport.Height = msg.Height - 3
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
				m.viewport.GotoBottom()
				// TODO: null-check store
				return m, SendChatMsg(prompt, m.store)
			}
		}
	case StoreMsg:
		m.store.Store = msg.Store
	case ChatMsg:
		m.chat = append(m.chat, msg)
		content := JoinChatMsg(m.chat)
		m.viewport.SetContent(content)
		m.viewport.GotoBottom()
	case ErrMsg: // critical error
		m.error = msg
		return m, tea.Quit
	}

	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Repl) View() string {
	if m.error != nil {
		return fmt.Sprintf("Error: %v\n", m.error)
	}

	return fmt.Sprintf(
		"%s\n%s\n",
		m.viewport.View(),
		m.input.View(),
	)
}

func Run(ctx context.Context, docs []schema.Document) error {

	model := New(ctx, docs)
	p := tea.NewProgram(
		model,
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
