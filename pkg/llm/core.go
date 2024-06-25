package llm

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/openai"
)

const (
	// OpenAI models
	GPT4o             = "gpt-4o"
	GPT4Turbo         = "gpt-4-turbo"
	GPT4Turbo0125     = "gpt-4-0125-preview"
	GPT4Turbo1106     = "gpt-4-1106-preview"
	GPT3Dot5Turbo0125 = "gpt-3.5-turbo-0125"
	GPT3Dot5Turbo1106 = "gpt-3.5-turbo-1106"

	// Anthropic models
	Claude35Sonnet = "claude-3-5-sonnet-20240620"
	Claude3Opus    = "claude-3-opus-20240229"
	Claude3Sonnet  = "claude-3-sonnet-20240229"
	Claude3Haiku   = "claude-3-haiku-20240307"
)

var Models = map[string]map[string]bool{
	"OpenAI": {
		GPT4o:             true,
		GPT4Turbo:         true,
		GPT4Turbo0125:     true,
		GPT4Turbo1106:     true,
		GPT3Dot5Turbo0125: true,
		GPT3Dot5Turbo1106: true,
	},
	"Anthropic": {
		Claude35Sonnet: true,
		Claude3Opus:    true,
		Claude3Sonnet:  true,
		Claude3Haiku:   true,
	},
}

func LookupModel(name string) (provider string, model string, found bool) {
	if name == "" {
		return "", "", false
	}

	for p, m := range Models {
		if _, ok := m[name]; ok {
			return p, name, true
		}
	}

	return "", "", false
}

func New(name string) (llms.Model, error) {
	if name == "" {
		return nil, fmt.Errorf("model name is empty")
	}

	provider, model, ok := LookupModel(name)
	if !ok {
		return nil, fmt.Errorf("unknown model %s", name)
	}

	switch provider {
	case "OpenAI":
		return openai.New(openai.WithModel(model))
	case "Anthropic":
		return anthropic.New(anthropic.WithModel(model))
	default:
		panic("unreachable")
	}

}

func CreateCompletion(ctx context.Context, llm llms.Model, prompt string, content string, stream bool) (string, error) {
	msgs := prepareMessages(prompt, content)

	options := []llms.CallOption{}

	if stream {
		streamFunc := func(ctx context.Context, chunk []byte) error {
			fmt.Print(string(chunk))
			return nil
		}

		options = append(options, llms.WithStreamingFunc(streamFunc))
	}

	resp, err := llm.GenerateContent(ctx, msgs, options...)
	if err != nil {
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return "", fmt.Errorf("empty response from model")
	}

	return choices[0].Content, nil

}

func prepareMessages(prompt string, content string) []llms.MessageContent {
	return []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},

		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: content}},
		},
	}
}
