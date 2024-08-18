package renderer

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

const (
	helpColor    = lipgloss.Color("#aaaaaa")
	errorColor   = lipgloss.Color("#cc3333")
	warningColor = lipgloss.Color("#ffcc00")
	successColor = lipgloss.Color("#46b946")
)

type Renderer struct {
	content *glamour.TermRenderer
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
	help    lipgloss.Style
}

func New(options ...glamour.TermRendererOption) *Renderer {
	contentRenderer, err := glamour.NewTermRenderer(options...)
	if err != nil {
		return nil
	}

	return &Renderer{
		content: contentRenderer,
		success: lipgloss.NewStyle().Foreground(successColor),
		warning: lipgloss.NewStyle().Foreground(warningColor),
		error:   lipgloss.NewStyle().Foreground(errorColor),
		help:    lipgloss.NewStyle().Foreground(helpColor).Italic(true),
	}
}

func (r *Renderer) RenderContent(in string) string {
	out, _ := r.content.Render(in)

	return out
}

func (r *Renderer) RenderSuccess(in string) string {
	return r.success.Render(in)
}

func (r *Renderer) RenderWarning(in string) string {
	return r.warning.Render(in)
}

func (r *Renderer) RenderError(in string) string {
	return r.error.Render(in)
}

func (r *Renderer) RenderHelp(in string) string {
	return r.help.Render(in)
}

// func (r *Renderer) RenderHelpMessage() string {
// 	help := "**Help**\n"
// 	help += "- `↑`/`↓` : navigate in history\n"
// 	help += "- `ctrl+l`: clear terminal but keep discussion history\n"
// 	help += "- `ctrl+c`: exit or interrupt command execution\n"

// 	return help
// }
