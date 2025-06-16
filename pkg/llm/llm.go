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
	O1              = "o1"
	O1Mini          = "o1-mini"
	O1Preview       = "o1-preview"
	O3Mini          = "o3-mini"
	O4Mini          = "o4-mini"
	GPT4Dot5Preview = "gpt-4.5-preview"
	GPT4Dot1        = "gpt-4.1"
	GPT4o           = "gpt-4o"
	GPT4oMini       = "gpt-4o-mini"
	GPT4            = "gpt-4"
	GPT4Turbo       = "gpt-4-turbo"
	ChatGPT4o       = "chatgpt-4o-latest"
	GPT3Dot5Turbo   = "gpt-3.5-turbo"

	// Anthropic models
	// https://docs.anthropic.com/en/docs/about-claude/models
	ClaudeOpus4       = "claude-opus-4-0"
	ClaudeSonnet4     = "claude-sonnet-4-0"
	ClaudeSonnet3Dot7 = "claude-3-7-sonnet-latest"
	ClaudeSonnet3Dot5 = "claude-3-5-sonnet-latest"
	ClaudeHaiku3Dot5  = "claude-3-5-haiku-latest"
	ClaudeOpus3       = "claude-3-opus-latest"

	// Google models
	// https://ai.google.dev/gemini-api/docs/models#model-variations
	Gemini2Dot5FlashPreview    = "gemini-2.5-flash-preview-05-20"
	Gemini2Dot5ProPreview      = "gemini-2.5-pro-preview-06-05"
	Gemini2Dot0Flash           = "gemini-2.0-flash"
	Gemini2Dot0FlashLite       = "gemini-2.0-flash-lite"
	Gemini2Dot0ProExperimental = "gemini-2.0-pro-exp-02-05"
	Gemini1Dot5Flash           = "gemini-1.5-flash"
	Gemini1Dot5Flash8B         = "gemini-1.5-flash-8b"
	Gemini1Dot5Pro             = "gemini-1.5-pro"
)

var hintTemplate = prompts.NewPromptTemplate(`
For the following content, focus on this aspect only: {{.hint}}
Note: If this focus is irrelevant to the content, disregard the focus.
Content: {{.content}}
`,
	[]string{"content", "hint"},
)

func New(name string) (llms.Model, error) {
	if name == "" {
		return nil, errors.New("model name is empty")
	}

	provider, model, ok := LookupModel(name)
	if !ok {
		return nil, fmt.Errorf("unsupported model: %s", name)
	}

	switch provider {
	case "openai":
		apiKey, err := env.OpenAIAPIKey()
		if err != nil {
			return nil, err
		}
		return openai.New(
			openai.WithModel(model),
			openai.WithToken(apiKey),
		)
	case "anthropic":
		apiKey, err := env.AnthropicAPIKey()
		if err != nil {
			return nil, err
		}
		return anthropic.New(
			anthropic.WithModel(model),
			anthropic.WithToken(apiKey),
		)
	case "google":
		apiKey, err := env.GeminiAPIKey()
		if err != nil {
			return nil, err
		}
		return googleai.New(
			context.Background(),
			googleai.WithAPIKey(apiKey),
			googleai.WithDefaultModel(model),
		)
	case "ollama":
		return ollama.New(
			ollama.WithModel(model),
			ollama.WithServerURL(env.OllamaHost()),
		)
	default:
		// TODO: check if provider is in connMap
		conn, ok := connMap[provider]
		if !ok {
			return nil, fmt.Errorf("unexpected error: provider %s not found", provider)
		}

		apiKey, err := env.Get(conn.GetEnvKey())
		if err != nil {
			return nil, err
		}
		return openai.New(
			openai.WithModel(model),
			openai.WithToken(apiKey),
			openai.WithBaseURL(conn.BaseURL),
		)
	}
}

func CreateCompletion(
	ctx context.Context,
	model llms.Model,
	writer io.Writer,
	msgs []llms.MessageContent,
	opts ...llms.CallOption,
) (err error) {
	// temporary workaround:
	// anthropic.generateMessagesContent() panics when it fails or returns no content
	// https://github.com/tmc/langchaingo/issues/993
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("stream completion panic: %v", r)
		}
	}()

	resp, err := model.GenerateContent(ctx, msgs, opts...)
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
	opts ...llms.CallOption,
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

	// extend the options with the streaming function
	extOpts := append([]llms.CallOption{}, opts...)
	extOpts = append(extOpts, llms.WithStreamingFunc(streamFunc))

	resp, err := model.GenerateContent(ctx, msgs, extOpts...)
	if err != nil {
		return fmt.Errorf("generate stream content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return errors.New("empty response from model")
	}

	return nil
}

func PrepareMessages(modelName string, prompt string, content string, hint string) []llms.MessageContent {
	altContent := content

	if hint != "" {
		// static template, so no need to check for error
		altContent, _ = hintTemplate.Format(map[string]any{"content": content, "hint": hint})
	}

	return []llms.MessageContent{
		{
			Role:  lookupSystemRole(modelName),
			Parts: []llms.ContentPart{llms.TextContent{Text: prompt}},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextContent{Text: altContent}},
		},
	}
}

// Exceptions where we need a different role for the system message.
//
// The role "system" has been deprecated in favor of "developer" for o1-family models provided by OpenAI.
// https://platform.openai.com/docs/api-reference/chat/create
//
// However, langchaingo does not support "developer" role yet.
// https://github.com/tmc/langchaingo/blob/0672790bb23a2c7e546a4a7aeffc9bef5bbd8c0b/llms/openai/openaillm.go#L60
var specialRoles = map[string]llms.ChatMessageType{
	"openai/o1":         llms.ChatMessageTypeGeneric,
	"openai/o1-mini":    llms.ChatMessageTypeGeneric,
	"openai/o1-preview": llms.ChatMessageTypeGeneric,
}

// lookupSystemRole returns the appropriate system role based on the model name.
func lookupSystemRole(modelName string) llms.ChatMessageType {
	// default role is "system"
	role := llms.ChatMessageTypeSystem

	if special, ok := specialRoles[modelName]; ok {
		role = special
	}

	return role
}
