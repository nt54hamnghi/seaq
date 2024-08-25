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

type components struct {
	prompt   input.Model
	renderer *renderer.Renderer
	spinner  *spinner.Spinner
}

type Repl struct {
	components
	engine *Engine                  // Language model engine
	store  vectorstores.VectorStore // Data store
	chat   Chat                     // Chat history
	error  error                    // Critical error
	ctx    context.Context
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
		r.engine.Model = model
		return nil
	}
}

func WithContext(ctx context.Context) ReplOption {
	return func(r *Repl) error {
		r.ctx = ctx
		return nil
	}
}

type ReplError struct {
	inner error
}

func (e ReplError) Error() string {
	return e.inner.Error()
}

func NewReplError(err error) ReplError {
	return ReplError{inner: err}
}

func NewRepl(docs []schema.Document, opts ...ReplOption) (*Repl, error) {
	if len(docs) == 0 {
		return nil, errors.New("no documents to load")
	}

	// initialize the repl
	repl := Repl{
		components: components{
			prompt: input.New(),
			renderer: renderer.New(
				// glamour.WithAutoStyle(),
				glamour.WithStandardStyle(renderer.DefaultStyle),
				glamour.WithWordWrap(100),
			),
			spinner: spinner.New(),
		},
		engine: &Engine{
			stream: make(chan StreamMsg),
		},
		chat: Chat{
			chat: make([]llms.MessageContent, 0),
		},
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
	if repl.engine.Model == nil {
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
	return tea.Batch(
		tea.ClearScreen,
		tea.Println(r.renderer.RenderHelpMessage()),
		textinput.Blink,
	)
}

func (r *Repl) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds       []tea.Cmd
		inputCmd   tea.Cmd
		spinnerCmd tea.Cmd
	)

	model, inputCmd := r.prompt.Update(msg)
	if prompt, ok := model.(input.Model); ok {
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
						r.engine.SendMessage(r.ctx, rawInput, &r.chat, r.store),
						r.engine.AwaitNext(),
					)
				}
			}
		}
	case StreamMsg:
		if msg.last {
			output := r.renderer.RenderContent(r.engine.buffer)
			r.chat.Append(newResponseMsg(output))
			r.engine.buffer = ""
			r.prompt.Focus()

			return r, tea.Sequence(
				tea.Println(output),
				textinput.Blink,
			)
		} else {
			// TODO: spinner should be stopped after the first message
			r.spinner.Stop()
			cmds = append(cmds, r.engine.AwaitNext())
		}
	case ChatError:
		output := r.renderer.RenderError(msg.Error())
		r.chat.Append(newResponseMsg(output))
		r.prompt.Focus()

		cmds = append(
			cmds,
			tea.Println(output),
			textinput.Blink,
		)
	case ReplError:
		r.error = msg
		return r, tea.Quit
	}

	return r, tea.Batch(cmds...)
}

func (r Repl) View() string {
	if r.error != nil {
		return r.renderer.RenderError(r.error.Error())
	}

	if r.spinner.Running() {
		return r.renderer.RenderContent(r.spinner.View() + "\n")
	}

	if len(r.engine.buffer) != 0 {
		return r.renderer.RenderContent(r.engine.buffer)
	}

	return r.prompt.View() + "\n"
}

func (r *Repl) Run() error {
	// f, err := tea.LogToFile("debug.log", "debug")
	// if err != nil {
	// 	return err
	// }
	// defer f.Close()

	p := tea.NewProgram(r,
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
