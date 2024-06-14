package openai

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

const (
	primingPrompt = `
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

func Prime(ctx context.Context, transcript string) (string, error) {
	return gen(ctx, primingPrompt, transcript)
}

func MakeConnection(ctx context.Context, transcript string) (string, error) {
	return gen(ctx, makeConnectionPrompt, transcript)
}

func gen(ctx context.Context, prompt string, transcript string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(apiKey)
	res, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: transcript,
			},
		},
	})

	if err != nil {
		return "", err
	}

	return res.Choices[0].Message.Content, nil
}
