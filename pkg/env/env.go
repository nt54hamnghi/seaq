package env

import (
	"fmt"
	"os"
	"strings"
)

// nolint: revive,gosec
const (
	OPENAI_API_KEY     = "OPENAI_API_KEY"
	ANTHROPIC_API_KEY  = "ANTHROPIC_API_KEY"
	GEMINI_API_KEY     = "GEMINI_API_KEY"
	YOUTUBE_API_KEY    = "YOUTUBE_API_KEY"
	CHROMA_URL         = "CHROMA_URL"
	X_AUTH_TOKEN       = "X_AUTH_TOKEN" // x.com
	X_CSRF_TOKEN       = "X_CSRF_TOKEN" // x.com
	UDEMY_ACCESS_TOKEN = "UDEMY_ACCESS_TOKEN"
	OLLAMA_HOST        = "OLLAMA_HOST"

	// seaq
	SEAQ_SUPPRESS_WARNINGS = "SEAQ_SUPPRESS_WARNINGS"
)

func Get(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%s is not set", key)
	}
	return val, nil
}

// OpenAIAPIKey returns the value of the OPENAI_API_KEY environment variable
// or an error if not set.
func OpenAIAPIKey() (string, error) {
	return Get(OPENAI_API_KEY)
}

// AnthropicAPIKey returns the value of the ANTHROPIC_API_KEY environment variable
// or an error if not set.
func AnthropicAPIKey() (string, error) {
	return Get(ANTHROPIC_API_KEY)
}

// GeminiAPIKey returns the value of the GEMINI_API_KEY environment variable
// or an error if not set.
func GeminiAPIKey() (string, error) {
	return Get(GEMINI_API_KEY)
}

// YoutubeAPIKey returns the value of the YOUTUBE_API_KEY environment variable
// or an error if not set.
func YoutubeAPIKey() (string, error) {
	return Get(YOUTUBE_API_KEY)
}

// ChromaURL returns the value of the CHROMA_URL environment variable
// or an error if not set.
func ChromaURL() (string, error) {
	return Get(CHROMA_URL)
}

// XAuthToken returns the value of the X_AUTH_TOKEN environment variable
// or an error if not set.
func XAuthToken() (string, error) {
	return Get(X_AUTH_TOKEN)
}

// XCSRFToken returns the value of the X_CSRF_TOKEN environment variable
// or an error if not set.
func XCSRFToken() (string, error) {
	return Get(X_CSRF_TOKEN)
}

// UdemyAccessToken returns the value of the UDEMY_ACCESS_TOKEN environment variable
// or an error if not set.
func UdemyAccessToken() (string, error) {
	return Get(UDEMY_ACCESS_TOKEN)
}

// SuppressWarnings returns the value of the SEAQ_SUPPRESS_WARNINGS environment variable
// or an error if not set.
func SuppressWarnings() bool {
	val, err := Get(SEAQ_SUPPRESS_WARNINGS)
	// show warnings by default if there's an error
	if err != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(val)) {
	// only suppress warnings when explicitly set to true
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

// OllamaHost returns the value of the OLLAMA_HOST environment variable
// or an error if not set.
func OllamaHost() string {
	host, err := Get(OLLAMA_HOST)
	if err != nil {
		return "http://localhost:11434"
	}
	return host
}
