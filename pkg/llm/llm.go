package llm

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
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

	// Google models
	Gemini15Flash = "gemini-1.5-flash"
	Gemini15Pro   = "gemini-1.5-pro"
	Gemini1Pro    = "gemini-1.0-pro"
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
	"Google": {
		Gemini15Flash: true,
		Gemini15Pro:   true,
		// FIXME: "gemini-1.0-pro" doesn't support system messages
		// https://ai.google.dev/gemini-api/docs/models/gemini#gemini-1.0-pro
		// Gemini1Pro:    true,
	},
}

// region: --- hint template

var hintTemplate = prompts.NewPromptTemplate(`
For the following content, focus on this aspect only: {{.hint}}
Note: If this focus is irrelevant to the content, disregard the focus.
Content: {{.content}}
`,
	[]string{"content", "hint"},
)

// endregion: --- hint template

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
		return nil, errors.New("model name is empty")
	}

	provider, model, ok := LookupModel(name)
	if !ok {
		return nil, fmt.Errorf("unsupported model: %s", name)
	}

	switch provider {
	case "OpenAI":
		apiKey, err := env.OpenAIAPIKey()
		if err != nil {
			return nil, err
		}
		return openai.New(
			openai.WithModel(model),
			openai.WithToken(apiKey),
		)
	case "Anthropic":
		apiKey, err := env.AnthropicAPIKey()
		if err != nil {
			return nil, err
		}
		return anthropic.New(
			anthropic.WithModel(model),
			anthropic.WithToken(apiKey),
		)
	case "Google":
		apiKey, err := env.GeminiAPIKey()
		if err != nil {
			return nil, err
		}
		return googleai.New(
			context.Background(),
			googleai.WithAPIKey(apiKey),
			googleai.WithDefaultModel(model),
		)
	default:
		panic("unreachable")
	}

}

func CreateCompletion(
	ctx context.Context,
	model llms.Model,
	writer io.Writer,
	msgs []llms.MessageContent,
) error {
	resp, err := model.GenerateContent(ctx, msgs)
	if err != nil {
		return err
	}

	if len(resp.Choices) == 0 {
		return errors.New("empty response from model")
	}

	_, err = io.WriteString(writer, resp.Choices[0].Content)
	if err != nil {
		return err
	}

	return nil
}

func CreateStreamCompletion(
	ctx context.Context,
	model llms.Model,
	writer io.Writer,
	msgs []llms.MessageContent,
) error {
	streamFunc := func(ctx context.Context, chunk []byte) error {
		_, err := writer.Write(chunk)
		return err
	}

	resp, err := model.GenerateContent(ctx, msgs,
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

func PrepareMessages(prompt string, content string, hint string) []llms.MessageContent {
	altContent := content

	if hint != "" {
		// static template, so no need to check for error
		altContent, _ = hintTemplate.Format(map[string]any{"content": content, "hint": hint})
	}

	return []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: altContent}},
		},
	}
}
