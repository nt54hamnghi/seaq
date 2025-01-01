package llm

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/anthropic"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
)

const (
	// OpenAI models
	// https://platform.openai.com/docs/models
	O1            = "o1"
	O1Mini        = "o1-mini"
	O1Preview     = "o1-preview"
	GPT4o         = "gpt-4o"
	GPT4oMini     = "gpt-4o-mini"
	GPT4          = "gpt-4"
	GPT4Turbo     = "gpt-4-turbo"
	ChatGPT4o     = "chatgpt-4o-latest"
	GPT3Dot5Turbo = "gpt-3.5-turbo"

	// Anthropic models
	// https://docs.anthropic.com/en/docs/about-claude/models
	Claude3Dot5Sonnet = "claude-3-5-sonnet-latest"
	Claude3Dot5Haiku  = "claude-3-5-haiku-latest"
	Claude3Opus       = "claude-3-opus-latest"
	Claude3Sonnet     = "claude-3-sonnet-20240229"
	Claude3Haiku      = "claude-3-haiku-20240307"

	// Google models
	// https://ai.google.dev/gemini-api/docs/models/gemini#model-variations
	Gemini2Dot0FlashExp = "gemini-2.0-flash-exp"
	Gemini1Dot5Flash    = "gemini-1.5-flash"
	Gemini1Dot5Flash8B  = "gemini-1.5-flash-8b"
	Gemini1Dot5Pro      = "gemini-1.5-pro"
)

const DefaultModel = Claude3Dot5Sonnet

var Models = map[string]map[string]bool{
	"OpenAI": {
		// O1:            true,
		// O1Mini:        true,
		// O1Preview:     true,
		GPT4o:         true,
		GPT4oMini:     true,
		GPT4:          true,
		GPT4Turbo:     true,
		ChatGPT4o:     true,
		GPT3Dot5Turbo: true,
	},
	"Anthropic": {
		Claude3Dot5Sonnet: true,
		Claude3Dot5Haiku:  true,
		Claude3Opus:       true,
		Claude3Sonnet:     true,
		Claude3Haiku:      true,
	},
	"Google": {
		Gemini2Dot0FlashExp: true,
		Gemini1Dot5Flash:    true,
		Gemini1Dot5Flash8B:  true,
		Gemini1Dot5Pro:      true,
	},
}

func init() {
	models, err := listOllamaLocalModels()
	if err != nil {
		// fmt.Printf("Warning: Failed to list Ollama models: %v", err)
		return
	}

	Models["Ollama"] = make(map[string]bool)
	for _, m := range models {
		Models["Ollama"][m] = true
	}
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
	case "Ollama":
		return ollama.New(
			ollama.WithModel(model),
			ollama.WithServerURL(env.OllamaHost()),
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
) (err error) {
	// temporary workaround:
	// anthropic.generateMessagesContent() panics when it fails or returns no content
	// https://github.com/tmc/langchaingo/issues/993
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("stream completion panic: %v", r)
		}
	}()

	resp, err := model.GenerateContent(ctx, msgs)
	if err != nil {
		return fmt.Errorf("generate content: %w", err)
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
) (err error) {
	// temporary workaround:
	// anthropic.generateMessagesContent() panics when it fails or returns no content
	// https://github.com/tmc/langchaingo/issues/993
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("stream completion panic: %v", r)
		}
	}()

	streamFunc := func(_ context.Context, chunk []byte) error {
		_, err := writer.Write(chunk)
		return err
	}

	resp, err := model.GenerateContent(ctx, msgs,
		llms.WithStreamingFunc(streamFunc),
	)
	if err != nil {
		return fmt.Errorf("generate stream content: %w", err)
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
