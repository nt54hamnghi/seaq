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

// RenderHelpMessage renders a help message
func (r *Renderer) RenderHelpMessage() string {
	help := "**Help**\n"
	help += "\n"
	help += "Keyboard shortcuts:\n"
	help += "- `↑`/`↓` : navigate in history\n"
	help += "- `ctrl+c`/`esc`: exit or interrupt command execution\n"
	help += "- `ctrl+h`: show help message\n"
	help += "\n"
	// help += "- `ctrl+l`: clear terminal but keep discussion history\n"
	help += "Commands:\n"
	help += "- `:q` or `:quit`: exit the program\n"

	return r.RenderContent(help)
}
