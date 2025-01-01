package env

import (
	"fmt"
	"os"
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
)

var globalEnvStore = NewEnvStore()

type Store struct {
	getEnv func(string) (string, error)
}

func NewEnvStore() *Store {
	return &Store{
		getEnv: func(s string) (string, error) {
			val, ok := os.LookupEnv(s)
			if !ok {
				return "", fmt.Errorf("%s is not set", s)
			}
			return val, nil
		},
	}
}

func (e *Store) Get(key string) (string, error) {
	return e.getEnv(key)
}

// OpenAIAPIKey returns the value of the OPENAI_API_KEY environment variable
// or an error if not set.
func OpenAIAPIKey() (string, error) {
	return globalEnvStore.Get(OPENAI_API_KEY)
}

// AnthropicAPIKey returns the value of the ANTHROPIC_API_KEY environment variable
// or an error if not set.
func AnthropicAPIKey() (string, error) {
	return globalEnvStore.Get(ANTHROPIC_API_KEY)
}

// GeminiAPIKey returns the value of the GEMINI_API_KEY environment variable
// or an error if not set.
func GeminiAPIKey() (string, error) {
	return globalEnvStore.Get(GEMINI_API_KEY)
}

// YoutubeAPIKey returns the value of the YOUTUBE_API_KEY environment variable
// or an error if not set.
func YoutubeAPIKey() (string, error) {
	return globalEnvStore.Get(YOUTUBE_API_KEY)
}

// ChromaURL returns the value of the CHROMA_URL environment variable
// or an error if not set.
func ChromaURL() (string, error) {
	return globalEnvStore.Get(CHROMA_URL)
}

// XAuthToken returns the value of the X_AUTH_TOKEN environment variable
// or an error if not set.
func XAuthToken() (string, error) {
	return globalEnvStore.Get(X_AUTH_TOKEN)
}

// XCSRFToken returns the value of the X_CSRF_TOKEN environment variable
// or an error if not set.
func XCSRFToken() (string, error) {
	return globalEnvStore.Get(X_CSRF_TOKEN)
}

// UdemyAccessToken returns the value of the UDEMY_ACCESS_TOKEN environment variable
// or an error if not set.
func UdemyAccessToken() (string, error) {
	return globalEnvStore.Get(UDEMY_ACCESS_TOKEN)
}

// OllamaHost returns the value of the OLLAMA_HOST environment variable
// or an error if not set.
func OllamaHost() string {
	host, err := globalEnvStore.Get(OLLAMA_HOST)
	if err != nil {
		return "http://localhost:11434"
	}
	return host
}
