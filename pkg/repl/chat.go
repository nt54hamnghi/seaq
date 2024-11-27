package repl

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/prompts"
	"github.com/tmc/langchaingo/vectorstores"
)

const defaultTemplate = `
Use the following context to answer the question at the end. Your answer should be simple and concise.

- If the context is irrelevant, state so explicitly and answer with the best of your knowledge.
- If you don't know the answer, just say that you don't know, don't try to make up an answer.
- If the question pertains to basic conversation (e.g., "hello", "thanks", etc.), respond as you would naturally in a conversation.

Format your answer in Markdown.

Context: {{.context}}

Question: {{.question}}

Helpful Answer:`

// region: --- error

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

// endregion: --- error

// region: --- helpers

func loadCombinedDocumentChain(model llms.Model) chains.StuffDocuments {
	promptTemplate := prompts.NewPromptTemplate(defaultTemplate, []string{"context", "question"})

	return chains.NewStuffDocuments(
		chains.NewLLMChain(model, promptTemplate),
	)
}

// endregion: --- helpers

// region: --- Engine

type StreamMsg struct {
	content string
	last    bool
}

type Chain struct {
	chains.ConversationalRetrievalQA
	buffer string
	stream chan StreamMsg
}

func NewChain(model llms.Model, store vectorstores.VectorStore) *Chain {
	conversation := memory.NewConversationBuffer(
		memory.WithChatHistory(memory.NewChatMessageHistory()),
	)

	return &Chain{
		ConversationalRetrievalQA: chains.NewConversationalRetrievalQA(
			loadCombinedDocumentChain(model),
			chains.LoadCondenseQuestionGenerator(model),
			vectorstores.ToRetriever(store, 5, vectorstores.WithScoreThreshold(0.7)),
			conversation,
		),
		buffer: "",
		stream: make(chan StreamMsg),
	}
}

func (c *Chain) AwaitNext() tea.Cmd {
	return func() tea.Msg {
		output := <-c.stream
		c.buffer += output.content
		return output
	}
}

func (c *Chain) run(ctx context.Context, question string) error {
	defer func() {
		c.stream <- StreamMsg{
			content: "",
			last:    true,
		}
	}()

	streamFunc := func(_ context.Context, chunk []byte) error {
		c.stream <- StreamMsg{
			content: string(chunk),
			last:    false,
		}
		return nil
	}

	resp, err := chains.Run(ctx, c, question,
		chains.WithStreamingFunc(streamFunc),
	)
	if err != nil {
		return err
	}
	if len(resp) == 0 {
		return errors.New("empty response from model")
	}

	return nil
}

func (c *Chain) SendMessage(ctx context.Context, question string) tea.Cmd {
	return func() tea.Msg {
		if err := c.run(ctx, question); err != nil {
			return NewChatError(err)
		}

		return nil
	}
}

// endregion: --- Engine

// func (ch *Chat) Debug_SendMessage(ctx context.Context, question string, model llms.Model, store vectorstores.VectorStore) tea.Cmd {
// 	return func() tea.Msg {
// 		return NewResponseChatMsg(question)
// 	}
// }
