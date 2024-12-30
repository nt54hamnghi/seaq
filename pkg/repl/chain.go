package repl

import (
	"context"
	"errors"
	"fmt"

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
- If the question pertains to basic conversation (e.g., "hello", "thanks", etc.), respond as you would naturally in a conversation. You don't need to explicitly point this out.

Format your answer in Markdown.

Context: {{.input_documents}}

Question: {{.question}}

Helpful Answer:`

// chatError is a recoverable error during chat
type chatError struct {
	inner error
}

// Error implements the error interface
func (e chatError) Error() string {
	return e.inner.Error()
}

// Unwrap implements the error interface
func (e chatError) Unwrap() error {
	return e.inner
}

// streamContentMsg contains the streaming content chunk from a chat operation
type streamContentMsg string

// streamEndMsg signals the end of a chat operation
type streamEndMsg struct{}

// chain is a history-aware conversational QA chain with document-retrieval capability.
// It's designed to interoperate with bubbletea.
type chain struct {
	// ConversationalRetrievalQA provides the core QA chain functionality
	chains.ConversationalRetrievalQA
	// buffer stores the accumulated response content
	buffer string
	// stream is used to send streaming chunks of the response
	stream chan tea.Msg
}

// newChain creates a new conversational QA Chain
// with the given language model and vector store.
func newChain(model llms.Model, store vectorstores.VectorStore) *chain {
	promptTemplate := prompts.NewPromptTemplate(
		defaultTemplate,
		[]string{"input_documents", "question"},
	)

	combineChain := chains.NewStuffDocuments(
		chains.NewLLMChain(model, promptTemplate),
	)

	condenseChain := chains.LoadCondenseQuestionGenerator(model)

	retriever := vectorstores.ToRetriever(store, 5,
		vectorstores.WithScoreThreshold(0.7),
	)

	memory := memory.NewConversationBuffer(
		memory.WithChatHistory(memory.NewChatMessageHistory()),
	)

	return &chain{
		ConversationalRetrievalQA: chains.NewConversationalRetrievalQA(
			combineChain,
			condenseChain,
			retriever,
			memory,
		),
		buffer: "",
		stream: make(chan tea.Msg),
	}
}

// start returns a bubbletea.Cmd that calls the chain with the given
// question and returns a tea.Msg.
func (c *chain) start(ctx context.Context, question string) tea.Cmd {
	return func() tea.Msg {
		if err := c.call(ctx, question); err != nil {
			return chatError{inner: err}
		}

		return nil
	}
}

// awaitNext returns a bubbletea.Cmd that reads the next streaming chunk,
// accumulates it in the buffer, and returns it as a tea.Msg.
func (c *chain) awaitNext() tea.Cmd {
	return func() tea.Msg {
		if c.stream == nil {
			// this is fatal, awaitNext should only be called after start
			return errors.New("unexpected nil stream: awaitNext called before start")
		}

		output := <-c.stream
		if msg, ok := output.(streamContentMsg); ok {
			c.buffer += string(msg)
		}
		return output
	}
}

// done signals the completion of a streaming response
// This method should be called when all streaming chunks have been sent.
func (c *chain) done() {
	if c.stream == nil {
		return
	}
	c.stream <- streamEndMsg{}
}

func (c *chain) call(ctx context.Context, question string) (err error) {
	// temporary workaround:
	// anthropic.generateMessagesContent() panics when it fails or returns no content
	// https://github.com/tmc/langchaingo/issues/993
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("completion panic, please try again")
		}

		c.done()
	}()

	streamFunc := func(_ context.Context, chunk []byte) error {
		c.stream <- streamContentMsg(chunk)
		return nil
	}

	res, err := chains.Call(ctx, c, map[string]any{"question": question},
		chains.WithStreamingFunc(streamFunc),
	)
	if err != nil {
		return fmt.Errorf("chain error: %w", err)
	}

	if _, ok := res[c.OutputKey]; !ok {
		return errors.New("chain returned no output")
	}

	return nil
}
