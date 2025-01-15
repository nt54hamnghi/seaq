package repl

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	spin "github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/nt54hamnghi/seaq/pkg/llm"
	"github.com/nt54hamnghi/seaq/pkg/rag"
	prompt "github.com/nt54hamnghi/seaq/pkg/repl/input"
	"github.com/nt54hamnghi/seaq/pkg/repl/renderer"
	"github.com/nt54hamnghi/seaq/pkg/repl/spinner"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

const defaultMargin = 10

var verticalMargin = defaultMargin

type ui struct {
	prompt   *prompt.Model
	renderer *renderer.Renderer
	spinner  *spinner.Spinner
	width    int
	height   int
}

type REPL struct {
	ui

	// main components
	model        llms.Model
	store        vectorstores.VectorStore
	chain        *chain
	conversation *conversation
	ctx          context.Context
	cancelFunc   context.CancelFunc

	// other options
	noStream bool
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

func WithContext(ctx context.Context) Option {
	return func(r *REPL) error {
		if ctx == nil {
			return errors.New("context is nil")
		}
		r.ctx = ctx
		return nil
	}
}

func WithNoStream(noStream bool) Option {
	return func(r *REPL) error {
		r.noStream = noStream
		return nil
	}
}

func defaultREPL() (*REPL, error) {
	store, err := rag.NewChromaStore()
	if err != nil {
		return nil, err
	}

	r := REPL{
		ui: ui{
			prompt:   prompt.New(),
			renderer: renderer.Default(),
			spinner:  spinner.New(),
		},
		store: store,
		ctx:   context.Background(),
	}

	return &r, nil
}

func New(name string, docs []schema.Document, opts ...Option) (*REPL, error) {
	if len(docs) == 0 {
		return nil, errors.New("no documents to load")
	}

	// initialize the REPL
	r, err := defaultREPL()
	if err != nil {
		return nil, err
	}

	// apply options, return on first error
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return nil, err
		}
	}

	// set model
	r.model, err = llm.New(name)
	if err != nil {
		return nil, err
	}

	// set conversation
	r.conversation = newConversation(name)

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

func (r *REPL) exit(err error) tea.Cmd {
	return func() tea.Msg {
		if err == nil {
			return tea.Quit
		}

		return tea.Sequence(
			tea.Printf("Error: %s\n", err.Error()),
			tea.Quit,
		)
	}
}

func (r *REPL) cancel() tea.Cmd {
	if r.cancelFunc == nil {
		output := r.renderer.RenderContent("Use Ctrl + d or /q to exit.")
		return tea.Sequence(
			tea.Println(r.prompt.Display()),
			tea.Println(output),
		)
	}
	r.spinner.Stop()
	r.cancelFunc()
	r.cancelFunc = nil
	r.chain.buffer = ""
	return nil
}

func (r *REPL) error(err error) tea.Cmd {
	return func() tea.Msg {
		return err
	}
}

func (r *REPL) help(_ []string) tea.Cmd {
	// TODO: check nil renderer

	return tea.Println(r.renderer.RenderHelpMessage())
}

func (r *REPL) save(args []string) tea.Cmd {
	// TODO: check nil conversation

	if len(args) == 0 {
		return r.conversation.saveJSON()
	}

	switch format := args[0]; format {
	case "json":
		return r.conversation.saveJSON()
	case "txt":
		return r.conversation.saveText()
	default:
		err := fmt.Errorf("invalid format: %s", format)
		return r.error(err)
	}
}

func parse(input string) (name string, args []string, err error) {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return "", nil, errors.New("empty input")
	}

	parts := strings.Fields(input)
	return parts[0], parts[1:], nil
}

func (r *REPL) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var promptCmd tea.Cmd
	r.prompt, promptCmd = r.prompt.Update(msg)

	cmds := []tea.Cmd{promptCmd}

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
		return r, promptCmd
	case spin.TickMsg:
		var spinnerCmd tea.Cmd
		if r.spinner.IsRunning() {
			r.spinner.Model, spinnerCmd = r.spinner.Update(msg)
			cmds = append(cmds, spinnerCmd)
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD, tea.KeyEsc:
			return r, tea.Quit
		case tea.KeyCtrlC:
			return r, tea.Sequence(r.cancel(), promptCmd)
		case tea.KeyEnter:
			displayCmd := tea.Println(r.prompt.Display())

			input := r.prompt.Value()
			r.prompt.Append(input)

			if input == "" {
				return r, tea.Sequence(displayCmd, promptCmd)
			}

			// ignore error because input is non-empty
			switch name, args, _ := parse(input); name {
			case "/?", "/help":
				return r, tea.Sequence(displayCmd, r.help(args), promptCmd)
			case "/s", "/save":
				return r, tea.Sequence(displayCmd, r.save(args))
			case "/c", "/clear":
				return r, tea.Sequence(tea.ClearScreen, promptCmd)
			case "/q", "/quit":
				return r, tea.Quit
			default:
				if strings.HasPrefix(name, "/") {
					err := fmt.Errorf("unknown command: %s. Type /? for help.", name) //nolint:revive
					return r, tea.Sequence(displayCmd, r.error(err), promptCmd)
				}

				r.spinner.Start()
				r.prompt.Blur()

				// ignore error because input is non-empty and role is always user
				_ = r.conversation.addMessage(input, roleUser)

				ctx, cancel := context.WithTimeout(r.ctx, 2*time.Minute)
				r.cancelFunc = cancel

				cmds = append(
					cmds,
					displayCmd,
					r.spinner.Tick,
					r.chain.start(ctx, input),
					r.chain.awaitNext(),
				)
			}
		}
	case saveConversationMsg:
		output := "Conversation saved to " + msg.path
		return r, tea.Sequence(
			tea.Println(r.renderer.RenderContent(output)),
			promptCmd,
		)
	case streamContentMsg:
		// In streaming mode, we want to stop the spinner immediately when content starts arriving
		// In non-streaming mode, we keep the spinner running until we get the complete response
		if !r.noStream {
			r.spinner.Stop()
		}
		cmds = append(cmds, r.chain.awaitNext())
	case streamEndMsg:
		// In non-streaming mode, we only stop the spinner once we have the complete response
		// In streaming mode, spinner was already stopped when content started arriving
		if r.noStream {
			r.spinner.Stop()
		}

		output := r.chain.buffer
		cmds := []tea.Cmd{}

		if output != "" {
			// ignore error because output is non-empty and role is always assistant
			_ = r.conversation.addMessage(output, roleAssistant)
			output = r.renderer.RenderContent(r.chain.buffer)
			verticalMargin = strings.Count(output, "\n") + defaultMargin

			cmds = append(cmds, tea.Println(output))
		}

		cmds = append(cmds, r.prompt.Focus())

		r.chain.buffer = ""
		return r, tea.Sequence(cmds...)
	case error:
		r.spinner.Stop()

		if errors.Is(msg, ErrNilStream) {
			return r, r.exit(msg)
		}

		output := msg.Error()

		if errors.Is(msg, context.Canceled) {
			output = "Operation cancelled."
		}

		return r, tea.Sequence(
			tea.Println(r.renderer.RenderError(output)),
			r.prompt.Focus(),
		)
	}

	return r, tea.Batch(cmds...)
}

func (r *REPL) View() string {
	if r.spinner.IsRunning() {
		// Add vertical padding (newlines) below the prompt and spinner to reserve space for the
		// streaming LLM response. The padding size is min(verticalMargin, availableHeight), where availableHeight
		// is terminal's height minus the spinner height and prompt height (2 lines).
		//
		// This padding is necessary because bubbletea's View() function can render incorrectly when there
		// isn't enough vertical space remaining. Without adequate padding, it may overwrite content
		// that was previously rendered using tea.Println(), leading to garbled output.
		spinner := r.renderer.RenderContent(r.spinner.View())
		spinnerHeight := strings.Count(spinner, "\n")
		margin := min(verticalMargin, r.height-(spinnerHeight+2))
		padding := strings.Repeat("\n", margin)
		return spinner + padding
	}

	// When streaming is enabled (noStream is false),
	// continuously update the view with the LLM's response as it's generated.
	if !r.noStream && len(r.chain.buffer) != 0 {
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
