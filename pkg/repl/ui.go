package repl

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nt54hamnghi/hiku/pkg/rag"
	"github.com/nt54hamnghi/hiku/pkg/repl/input"
	"github.com/nt54hamnghi/hiku/pkg/repl/renderer"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type UiComponents struct {
	prompt   input.Model
	renderer *renderer.Renderer
}

type Repl struct {
	UiComponents
	model llms.Model               // Language model
	store vectorstores.VectorStore // Data store
	chat  []ChatMsg                // Chat history
	error error                    // Critical error
	ctx   context.Context
}

type ReplOption func(*Repl) error

func WithStore(store vectorstores.VectorStore) ReplOption {
	return func(r *Repl) error {
		r.store = store
		return nil
	}
}

func WithDefaultStore() ReplOption {
	return func(r *Repl) (err error) {
		r.store, err = rag.NewChromaStore()
		if err != nil {
			return err
		}
		return nil
	}
}

func WithModel(model llms.Model) ReplOption {
	return func(r *Repl) error {
		r.model = model
		return nil
	}
}

func WithContext(ctx context.Context) ReplOption {
	return func(r *Repl) error {
		r.ctx = ctx
		return nil
	}
}

func NewRepl(docs []schema.Document, opts ...ReplOption) (*Repl, error) {
	if len(docs) == 0 {
		return nil, errors.New("no documents to load")
	}

	// initialize the repl
	repl := Repl{
		UiComponents: UiComponents{
			prompt: input.New(),
			renderer: renderer.New(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(100),
			),
		},
		chat: make([]ChatMsg, 0),
	}

	// apply options, if any
	for _, opt := range opts {
		err := opt(&repl)
		// return as soon as an error is encountered
		if err != nil {
			return nil, err
		}
	}

	// set the default context if not provided
	if repl.ctx == nil {
		repl.ctx = context.Background()
	}

	// refuse to proceed without a model
	// TODO: might want to use a default model
	if repl.model == nil {
		return nil, errors.New("no model provided")
	}

	// set the default store if not provided
	if repl.store == nil {
		err := WithDefaultStore()(&repl)
		if err != nil {
			return nil, err
		}
	}

	// add documents to the store
	_, err := repl.store.AddDocuments(repl.ctx, docs)
	if err != nil {
		return nil, err
	}

	return &repl, nil
}

func (r Repl) Init() tea.Cmd {
	return tea.ClearScreen
}

func (r Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds     []tea.Cmd
		inputCmd tea.Cmd
	)

	model, inputCmd := r.prompt.Update(msg)
	if prompt, ok := model.(input.Model); ok {
		r.prompt = prompt
	}
	cmds = append(cmds, inputCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.prompt.Width = msg.Width / 3 * 2
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return r, tea.Quit
		case tea.KeyEnter:
			switch strings.ToLower(r.prompt.Value()) {
			case "exit", "bye":
				return r, tea.Quit
			default:
				rawInput := r.prompt.Value()

				if rawInput != "" {
					input := r.prompt.AsString()
					r.prompt.Append(rawInput)
					r.prompt.Blur()
					cmds = append(
						cmds,
						tea.Println(input),
						SendChatMsg(rawInput, r.model, r.store),
					)
				}
			}
		}
	case ChatMsg:
		r.chat = append(r.chat, msg)
		output := r.renderer.RenderContent(msg.Answer)
		r.prompt.Focus()
		cmds = append(
			cmds,
			tea.Println(output),
		)
	case ErrMsg: // critical error
		r.error = msg
		return r, tea.Quit
	}

	return r, tea.Batch(cmds...)
}

func (r Repl) View() string {
	if r.error != nil {
		return fmt.Sprintf("Error: %v\n", r.error)
	}

	return r.prompt.View() + "\n"
}

func (r *Repl) Run() error {
	p := tea.NewProgram(r,
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

type ErrMsg struct {
	error
}

func (e ErrMsg) Error() string {
	return e.error.Error()
}

func (e ErrMsg) Unwrap() error {
	return e.error
}
