package repl

import (
	"context"
	"errors"
	"strings"

	spin "github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nt54hamnghi/hiku/pkg/rag"
	"github.com/nt54hamnghi/hiku/pkg/repl/input"
	"github.com/nt54hamnghi/hiku/pkg/repl/renderer"
	"github.com/nt54hamnghi/hiku/pkg/repl/spinner"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

type ui struct {
	prompt   *input.Model
	renderer *renderer.Renderer
	spinner  *spinner.Spinner
}

type REPL struct {
	ui
	model llms.Model
	store vectorstores.VectorStore
	chain *chain
	error error // Critical error
	ctx   context.Context
}

type Option func(*REPL) error

func WithDefaultStore() Option {
	return func(r *REPL) (err error) {
		if r.store, err = rag.NewChromaStore(); err != nil {
			return err
		}
		return nil
	}
}

func WithStore(store vectorstores.VectorStore) Option {
	return func(r *REPL) error {
		r.store = store
		return nil
	}
}

func WithModel(model llms.Model) Option {
	return func(r *REPL) error {
		r.model = model
		return nil
	}
}

func WithContext(ctx context.Context) Option {
	return func(r *REPL) error {
		r.ctx = ctx
		return nil
	}
}

func New(docs []schema.Document, opts ...Option) (*REPL, error) {
	if len(docs) == 0 {
		return nil, errors.New("no documents to load")
	}

	// initialize the r
	r := REPL{
		ui: ui{
			prompt:   input.New(),
			renderer: renderer.Default(),
			spinner:  spinner.New(),
		},
	}

	// apply options, if any
	// return as soon as an error is encountered
	for _, opt := range opts {
		if err := opt(&r); err != nil {
			return nil, err
		}
	}

	if r.ctx == nil {
		return nil, errors.New("context is nil")
	}

	if r.model == nil {
		return nil, errors.New("model is nil")
	}

	if r.store == nil {
		return nil, errors.New("store is nil")
	}

	// add documents to the store
	if _, err := r.store.AddDocuments(r.ctx, docs); err != nil {
		return nil, err
	}

	// initialize the chain
	r.chain = newChain(r.model, r.store)

	return &r, nil
}

func (r REPL) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		tea.Println(r.renderer.RenderHelpMessage()),
		textinput.Blink,
	)
}

func (r *REPL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		inputCmd   tea.Cmd
		spinnerCmd tea.Cmd
	)

	model, inputCmd := r.prompt.Update(msg)
	if prompt, ok := model.(*input.Model); ok {
		r.prompt = prompt
	}
	cmds = append(cmds, inputCmd)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width / 3 * 2

		r.renderer = renderer.New(
			// glamour.WithAutoStyle(), // bug: this leaks style string to the input field
			glamour.WithStandardStyle(renderer.DefaultStyle),
			glamour.WithWordWrap(w),
		)

		r.prompt.Width = w
		r.prompt.SetValue("")
		r.prompt.Reset()

		return r, nil
	case spin.TickMsg:
		if r.spinner.Running() {
			r.spinner.Model, spinnerCmd = r.spinner.Update(msg)
			cmds = append(cmds, spinnerCmd)
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return r, tea.Quit
		case tea.KeyCtrlH:
			cmds = append(
				cmds,
				tea.Println(r.renderer.RenderHelpMessage()),
				textinput.Blink,
			)
		case tea.KeyEnter:
			switch strings.ToLower(r.prompt.Value()) {
			case ":q", ":quit":
				// if cb := r.chain.Memory.(*memory.ConversationBuffer); cb != nil {
				// 	log.Println(cb.ChatHistory.Messages(context.Background()))
				// }
				return r, tea.Quit
			default:
				rawInput := r.prompt.Value()

				if rawInput != "" {
					r.spinner.Start()

					input := r.prompt.AsString()
					r.prompt.Append(rawInput)
					r.prompt.Blur()

					cmds = append(
						cmds,
						tea.Println(input),
						r.spinner.Tick, // advance spinner
						r.chain.run(r.ctx, rawInput),
						r.chain.awaitNext(),
					)
				}
			}
		}
	case streamMsg:
		if msg.last {
			output := r.renderer.RenderContent(r.chain.buffer)
			r.chain.buffer = ""
			r.prompt.Focus()

			return r, tea.Sequence(
				tea.Println(output),
				textinput.Blink,
			)
		} else {
			// TODO: spinner should be stopped after the first message
			r.spinner.Stop()
			cmds = append(cmds, r.chain.awaitNext())
		}
	case chatError:
		output := r.renderer.RenderError(msg.Error())
		r.prompt.Focus()

		cmds = append(
			cmds,
			tea.Println(output),
			textinput.Blink,
		)
	case error:
		r.error = msg
		return r, tea.Quit
	}

	return r, tea.Batch(cmds...)
}

func (r REPL) View() string {
	if r.error != nil {
		return r.renderer.RenderError(r.error.Error())
	}

	if r.spinner.Running() {
		return r.renderer.RenderContent(r.spinner.View() + "\n")
	}

	if len(r.chain.buffer) != 0 {
		return r.renderer.RenderContent(r.chain.buffer)
	}

	return r.prompt.View() + "\n"
}

func (r *REPL) Run() error {
	// debug()

	p := tea.NewProgram(r,
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// nolint:unused
func debug() error {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
