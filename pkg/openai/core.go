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

const (
	PrimingPrompt = `
Act as a creative, passionate, and knowledgeable educator.

You embrace the idea of priming the mind before learning, which involves reviewing key terms, ideas, and relationships before learning a new skill or complex topic.

Your task is to read through a transcript provided by a student carefully and create a helpful overview of the video based on this transcript. The overview should be of medium length and cover the main concepts from the video. Please add explanations for terminologies and include helpful examples. For programming related topics, you can include code snippets to illustrate the concepts.

The transcript is typically from a YouTube video, and the student is a novice at the topic covered in the video.

The response should be in Markdown format with a title. 	
`
	makeConnectionPrompt = `
Act as a creative, passionate, and knowledgeable educator.

You embrace the idea of promoting connections between indirect concepts of a topic, which forces the brain to explore multiple paths and perspectives that could reinforce your understanding.

Your task is to read through a transcript provided by a student carefully and create a list of 4 non-obvious questions about the topic covered in the transcript. Each question should have a concise answer in a bullet-point format, and please use simple language for the answer. For each answer, please explore one ideas (in the form of a bullet point) related to the topic but not mentioned in the transcript.

The transcript is typically from a YouTube video, and the student is a novice at the topic covered in the video.

The response should be in Markdown format with a title.  
`
)

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
