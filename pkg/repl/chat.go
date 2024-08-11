package repl

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	prefix = "> "
)

type ChatMsg struct {
	Prompt string
	Answer string
}

// FIXME: use pointer receiver since a chat message can be large
func (c ChatMsg) String() string {
	return prefix + c.Prompt + "\n" + c.Answer
}

func JoinChatMsg(chat []ChatMsg) string {
	var sb strings.Builder

	for _, c := range chat {
		sb.WriteString(c.String())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func SendChatMsg(prompt string) tea.Cmd {
	return func() tea.Msg {
		// TODO: make request to API

		return ChatMsg{
			Prompt: prompt,
			Answer: prompt, // echo for now
		}
	}
}
