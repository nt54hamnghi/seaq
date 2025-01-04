package repl

import (
	"context"
	"errors"
	"strings"

	spin "github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/nt54hamnghi/seaq/pkg/rag"
	"github.com/nt54hamnghi/seaq/pkg/repl/input"
	"github.com/nt54hamnghi/seaq/pkg/repl/renderer"
	"github.com/nt54hamnghi/seaq/pkg/repl/spinner"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

const defaultMargin = 10

var verticalMargin = defaultMargin

type ui struct {
	prompt   *input.Model
	renderer *renderer.Renderer
	spinner  *spinner.Spinner
	width    int
	height   int
}

type REPL struct {
	ui
	model llms.Model
	store vectorstores.VectorStore
	chain *chain
	ctx   context.Context
}

type Option func(*REPL) error

func WithStore(store vectorstores.VectorStore) Option {
	return func(r *REPL) error {
		if store == nil {
			return errors.New("store is nil")
		}
		r.store = store
		return nil
	}
}

func WithModel(model llms.Model) Option {
	return func(r *REPL) error {
		if model == nil {
			return errors.New("model is nil")
		}
		r.model = model
		return nil
	}
}

func WithContext(ctx context.Context) Option {
	return func(r *REPL) error {
		if ctx == nil {
			return errors.New("context is nil")
		}
		r.ctx = ctx
		return nil
	}
}

func Default() (*REPL, error) {
	store, err := rag.NewChromaStore()
	if err != nil {
		return nil, err
	}

	model, err := llm.New(llm.DefaultModel)
	if err != nil {
		return nil, err
	}

	r := REPL{
		ui: ui{
			prompt:   input.New(),
			renderer: renderer.Default(),
			spinner:  spinner.New(),
		},
		model: model,
		store: store,
		ctx:   context.Background(),
	}

	return &r, nil
}

func New(docs []schema.Document, opts ...Option) (*REPL, error) {
	if len(docs) == 0 {
		return nil, errors.New("no documents to load")
	}

	// initialize the REPL
	r, err := Default()
	if err != nil {
		return nil, err
	}

	// apply options, return on first error
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	// add documents to the store
	if _, err := r.store.AddDocuments(r.ctx, docs); err != nil {
		return nil, err
	}

	// initialize the chain
	r.chain = newChain(r.model, r.store)

	return r, nil
}

func (r *REPL) Init() tea.Cmd {
	return tea.Batch(
		tea.ClearScreen,
		textinput.Blink,
	)
}

func (r *REPL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var inputCmd tea.Cmd
	r.prompt, inputCmd = r.prompt.Update(msg)

	cmds := []tea.Cmd{inputCmd}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.height = msg.Height

		w := r.width / 3 * 2
		r.prompt.Width = w
		r.renderer = renderer.New(
			glamour.WithStandardStyle(renderer.DefaultStyle),
			glamour.WithWordWrap(w),
		)
		return r, inputCmd
	case spin.TickMsg:
		var spinnerCmd tea.Cmd
		if r.spinner.IsRunning() {
			r.spinner.Model, spinnerCmd = r.spinner.Update(msg)
			cmds = append(cmds, spinnerCmd)
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return r, tea.Quit
		case tea.KeyEnter:
			inputValue := r.prompt.Value()
			inputDisplay := r.prompt.Display()
			r.prompt.Append(inputValue)

			switch strings.ToLower(inputValue) {
			case "":
				cmds = append(cmds, tea.Println(inputDisplay))
			case "/?", "/help":
				return r, tea.Sequence(
					tea.Println(inputDisplay),
					tea.Println(r.renderer.RenderHelpMessage()),
					inputCmd,
				)
			case "/c", "/clear":
				return r, tea.Sequence(tea.ClearScreen, inputCmd)
			case "/q", "/quit":
				return r, tea.Quit
			default:
				r.spinner.Start()
				r.prompt.Blur()

				cmds = append(
					cmds,
					tea.Println(inputDisplay),
					r.spinner.Tick,
					r.chain.start(r.ctx, inputValue),
					r.chain.awaitNext(),
				)
			}
		}

	case streamContentMsg:
		r.spinner.Stop()
		cmds = append(cmds, r.chain.awaitNext())
	case streamEndMsg:
		output := r.chain.buffer
		cmds := []tea.Cmd{}

		if output != "" {
			output = r.renderer.RenderContent(r.chain.buffer)
			verticalMargin = strings.Count(output, "\n") + defaultMargin
			cmds = append(cmds, tea.Println(output))
		}

		cmds = append(cmds,
			r.prompt.Focus(),
			textinput.Blink,
		)

		r.chain.buffer = ""
		return r, tea.Sequence(cmds...)
	case error:
		output := msg.Error()
		r.spinner.Stop()

		if errors.Is(msg, ErrNilStream) {
			return r, tea.Sequence(
				tea.Printf("Error: %s\n", output),
				tea.Quit,
			)
		}

		return r, tea.Sequence(
			tea.Println(r.renderer.RenderError(output)),
			r.prompt.Focus(),
			textinput.Blink,
		)
	}

	return r, tea.Batch(cmds...)
}

func (r *REPL) View() string {
	if r.spinner.IsRunning() {
		spinner := r.renderer.RenderContent(r.spinner.View())

		spinnerHeight := strings.Count(spinner, "\n")
		margin := min(verticalMargin, r.height-(spinnerHeight+2))
		padding := strings.Repeat("\n", margin)

		return spinner + padding
	}

	if len(r.chain.buffer) != 0 {
		return r.renderer.RenderContent(r.chain.buffer)
	}

	return r.prompt.View() + "\n"
}

func (r *REPL) Run() error {
	p := tea.NewProgram(r)
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

// nolint:unused
func _debug() error {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
