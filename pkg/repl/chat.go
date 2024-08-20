package repl

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/schema"
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
3. Keep your response concise and directly relevant to the query.
4. If the context is insufficient:
   a. Provide as much information as you can from the context.
   b. Supplement with your general knowledge, and be explicit if you do so.
5. If the context is empty or irrelevant, answer the query to the best of your ability.
6. Suggest other concepts/ideas that could be related to the query or context, and be explicit if you do so.
7. Ensure your response is fluent and coherent and avoid simply listing facts.
`

// region: --- helpers

// TODO: use langchaingo built-in template functions
func constructPrompt(question string, tempate string, docs []schema.Document) string {
	context := ""
	for _, d := range docs {
		context += d.PageContent + "\n"
	}
	return fmt.Sprintf(tempate, context, question)
}

func newPromptMsg(prompt string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
	}
}

func newResponseMsg(response string) llms.MessageContent {
	return llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{llms.TextContent{Text: response}},
	}
}

// endregion: --- helpers

// recoverable error during chat
type ChatError struct {
	inner error
}

func (e ChatError) Error() string {
	return e.inner.Error()
}

func NewChatError(err error) ChatError {
	return ChatError{inner: err}
}

// region: --- Chat

type Chat struct {
	chat []llms.MessageContent
}

func (ch *Chat) Append(msg llms.MessageContent) {
	ch.chat = append(ch.chat, msg)
}

func (ch Chat) Len() int {
	return len(ch.chat)
}

func (ch Chat) GetContents() []llms.MessageContent {
	return ch.chat
}

// endregion: --- Chat

// region: --- Engine

type StreamMsg struct {
	content string
	last    bool
}

type Engine struct {
	llms.Model
	stream chan StreamMsg
	buffer string
}

// TODO: Change name
func (e *Engine) startCompletion(ctx context.Context, msgs []llms.MessageContent) error {
	defer func() {
		e.stream <- StreamMsg{
			content: "",
			last:    true,
		}
	}()

	streamFunc := func(ctx context.Context, chunk []byte) error {
		e.stream <- StreamMsg{
			content: string(chunk),
			last:    false,
		}
		return nil
	}

	resp, err := e.GenerateContent(ctx, msgs,
		llms.WithStreamingFunc(streamFunc),
	)

	if err != nil {
		return err
	}

	if len(resp.Choices) == 0 {
		return errors.New("empty response from model")
	}

	return nil
}

func (e *Engine) AwaitNext() tea.Cmd {
	return func() tea.Msg {
		output := <-e.stream
		e.buffer += output.content
		return output
	}
}

func (e *Engine) SendMessage(ctx context.Context, question string, chat *Chat, store vectorstores.VectorStore) tea.Cmd {
	return func() tea.Msg {
		retrievedDocs, err := store.SimilaritySearch(
			ctx, question, 10,
			vectorstores.WithScoreThreshold(0.7),
		)
		if err != nil {
			return NewChatError(err)
		}

		prompt := constructPrompt(question, template, retrievedDocs)
		chat.Append(newPromptMsg(prompt))

		err = e.startCompletion(ctx, chat.GetContents())
		if err != nil {
			return NewChatError(err)
		}

		chat.chat[chat.Len()-1] = newPromptMsg(question)

		return nil
	}
}

// endregion: --- Engine

// func (ch *Chat) Debug_Query(ctx context.Context, question string, model llms.Model, store vectorstores.VectorStore) tea.Cmd {
// 	return func() tea.Msg {
// 		return NewResponseChatMsg(question)
// 	}
// }
