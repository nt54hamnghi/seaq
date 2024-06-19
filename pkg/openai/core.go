package openai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const OPENAI_API_KEY = "OPENAI_API_KEY"

func CreateCompletionStream(ctx context.Context, prompt string, content string) error {
	apiKey := os.Getenv(OPENAI_API_KEY)
	client := openai.NewClient(apiKey)

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: prepareMessages(prompt, content),
		Stream:   true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("chat completion error: %w", err)
	}

	for {
		res, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("streaming error: %w", err)
		}

		fmt.Printf(res.Choices[0].Delta.Content)
		time.Sleep(30 * time.Millisecond)
	}

}

func CreateCompletion(ctx context.Context, prompt string, content string) error {
	apiKey := os.Getenv(OPENAI_API_KEY)
	client := openai.NewClient(apiKey)

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: prepareMessages(prompt, content),
	}

	res, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("chat completion error: %w", err)
	}

	fmt.Println(res.Choices[0].Message.Content)
	return nil
}

func prepareMessages(prompt string, content string) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: prompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: content,
		},
	}
}
