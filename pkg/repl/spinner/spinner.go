package spinner

import (
	"github.com/charmbracelet/bubbles/spinner"
)

type Spinner struct {
	spinner.Model
	running bool
	message string
}

func New() *Spinner {
	return &Spinner{
		Model: spinner.New(
			spinner.WithSpinner(spinner.Points),
		),
		running: false,
		message: "Thinking",
	}
}

func (s Spinner) Running() bool {
	return s.running
}

func (s *Spinner) Start() {
	if !s.running {
		s.running = true
	}
}

func (s *Spinner) Stop() {
	if s.running {
		s.running = false
	}
}

func (s Spinner) View() string {
	if !s.running {
		return ""
	}
	return s.message + " " + s.Model.View()
}
