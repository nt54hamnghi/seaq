package renderer

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpColor    = lipgloss.Color("#aaaaaa")
	errorColor   = lipgloss.Color("#f38ba8")
	warningColor = lipgloss.Color("#f9e2af")
	successColor = lipgloss.Color("#89b4fa")

	errorPrefix   = "Error:"
	warningPrefix = "Warning:"
	successPrefix = "Success:"

	DefaultStyle = styles.DarkStyle
)

// region: --- helpers

func renderMessage(msg string, style lipgloss.Style, prefix string) string {
	if prefix == "" {
		prefix = "Message:"
	}
	msg = fmt.Sprintf("\n%s %s\n", prefix, msg)
	return style.Render(msg)
}

// endregion: --- helpers

type Renderer struct {
	*glamour.TermRenderer // markdown renderer
	success               lipgloss.Style
	warning               lipgloss.Style
	error                 lipgloss.Style
	help                  lipgloss.Style
}

func New(options ...glamour.TermRendererOption) *Renderer {
	contentRenderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return nil
	}

	return &Renderer{
		TermRenderer: contentRenderer,
		success:      lipgloss.NewStyle().Foreground(successColor),
		warning:      lipgloss.NewStyle().Foreground(warningColor),
		error:        lipgloss.NewStyle().Foreground(errorColor),
		help:         lipgloss.NewStyle().Foreground(helpColor).Italic(true),
	}
}

func Default() *Renderer {
	return New(
		glamour.WithStandardStyle(DefaultStyle),
		glamour.WithWordWrap(100),
	)
}

// RenderContent renders the content using the glamour renderer
func (r *Renderer) RenderContent(content string) string {
	out, _ := r.Render(content)
	return out
}

// RenderSuccess renders a success message
func (r *Renderer) RenderSuccess(msg string) string {
	return renderMessage(msg, r.success, successPrefix)
}

// RenderWarning renders a warning message
func (r *Renderer) RenderWarning(msg string) string {
	return renderMessage(msg, r.warning, warningPrefix)
}

// RenderError renders an error message
func (r *Renderer) RenderError(msg string) string {
	return renderMessage(msg, r.error, errorPrefix)
}

const helpMessage = `**Commands:**
- /?, /help  : Show help message
- /q, /quit  : Exit the program
- /c, /clear : Clear the terminal

**Keyboard Shortcuts:**
- ↑/↓        : Navigate in history
- ctrl+h     : Show help message
- ctrl+l     : Clear the terminal
- ctrl+c/esc : Exit or interrupt command execution
`

func (r *Renderer) RenderHelpMessage() string {
	return r.RenderContent(helpMessage)
}
