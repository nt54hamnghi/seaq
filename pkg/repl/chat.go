package repl

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nt54hamnghi/hiku/pkg/llm"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/vectorstores"
)

const template = `
Answer the question based on the provided context. 

Context:
{%s}

Query:
{%s}

Instructions:
1. First, analyze the context thoroughly.
2. Attempt to answer the query using only the information provided in the context.
3. If the context is insufficient:
   a. Provide as much information as you can from the context.
   b. Supplement with your general knowledge, and ONLY be explicit if you do so.
4. Keep your response concise and directly relevant to the query.
`

type ChatMsg struct {
	Prompt string
	Answer string
	Error  error
}

// FIXME: use pointer receiver since a chat message can be large
func (c ChatMsg) String() string {
	if c.Error != nil {
		return c.Error.Error()
	}

	return c.Answer
}

func JoinChatMsg(chat []ChatMsg) string {
	var sb strings.Builder

	for _, c := range chat {
		sb.WriteString(c.String())
		sb.WriteString("\n\n")
	}

	return sb.String()
}

func SendChatMsg(question string, model llms.Model, store vectorstores.VectorStore) tea.Cmd {
	return func() tea.Msg {
		ctx := context.TODO()

		retrievedDocs, err := store.SimilaritySearch(
			ctx, question, 3,
			vectorstores.WithScoreThreshold(0.7),
		)
		if err != nil {
			return ChatMsg{
				Prompt: question,
				Error:  err,
			}
		}

		context := ""
		for _, d := range retrievedDocs {
			context += d.PageContent + "\n"
		}

		var buf bytes.Buffer
		prompt := fmt.Sprintf(template, context, question)

		llm.CreateStreamCompletion(ctx, model, []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
			},
		}, &buf)

		return ChatMsg{
			Prompt: question,
			Answer: buf.String(),
		}
	}
}

func Debug_SendChatMsg(prompt string, store vectorstores.VectorStore) tea.Cmd {
	return func() tea.Msg {

		// answer := ""
		// for i := 0; i < 100; i++ {
		// 	answer += fmt.Sprintf("%d\n", i)
		// }

		return ChatMsg{
			Prompt: prompt,
			Answer: prompt,
		}
	}
}
