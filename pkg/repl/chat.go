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
You are a knowledgeable AI assistant tasked with answering questions based on provided context. Follow these instructions carefully:

Context:
%s

Query:
%s

Instructions:
Analyze the context and query, then provide a response in Markdown format with the following sections:

1. Answer:
	- Thoroughly examine the context.
	- Provide a SHORT AND CONCISE answer using only the provided context.
	- If the context is insufficient:
		- Include as much relevant information as possible from the context.
		- Supplement with your knowledge, clearly stating when you do so.
	- If the context is empty or irrelevant:
		- Clearly state so and provide an answer based on your knowledge.
	- Ensure your response is fluent, coherent, and not just a list of facts.

2. Additional Information:
	- Suggest 1-2 related concepts or ideas that are related to the query or context and may interest the user.
	- Briefly explain how each concept relates to the original query.


3. Relevant Ideas:
   - If the context includes terms that are unclear or insufficiently explained, provide a brief explanation with your own knowledge.
     - For instance, if the context mentions NAT gateways, explain what NAT is and how it pertains to networking.
     - The explanation can extend beyond the immediate context.
   - Suggest 1-2 additional ideas related to the query that were not explicitly mentioned in the context.

Remember to maintain a helpful, informative tone throughout your response. 
If you're unsure about any information, state your uncertainty clearly.
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
