package repl

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nt54hamnghi/hiku/pkg/rag"
	"github.com/tmc/langchaingo/vectorstores"
)

type ChatMsg struct {
	Prompt string
	Answer string
	Error  error
}

// FIXME: use pointer receiver since a chat message can be large
func (c ChatMsg) String() string {
	if c.Error != nil {
		return c.Error.Error() + "\n"
	}

	return c.Answer + "\n"
}

func JoinChatMsg(chat []ChatMsg) string {
	var sb strings.Builder

	for _, c := range chat {
		sb.WriteString(c.String())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func SendChatMsg(prompt string, store *rag.DocumentStore) tea.Cmd {
	return func() tea.Msg {
		results, err := store.SimilaritySearch(context.TODO(), prompt, 2, vectorstores.WithScoreThreshold(0.7))
		if err != nil {
			return ChatMsg{
				Prompt: prompt,
				Error:  err,
			}
		}

		answer := "Nothing found"
		if len(results) >= 1 {
			answer = results[0].PageContent
		}

		// TODO: forward data to LLM models

		return ChatMsg{
			Prompt: prompt,
			Answer: answer,
		}
	}
}

// func Debug_SendChatMsg(prompt string, store *rag.DocumentStore) tea.Cmd {
// 	return func() tea.Msg {

// 		answer := ""
// 		for i := 0; i < 100; i++ {
// 			answer += fmt.Sprintf("%d\n", i)
// 		}

// 		return ChatMsg{
// 			Prompt: prompt,
// 			Answer: answer,
// 		}
// 	}
// }
