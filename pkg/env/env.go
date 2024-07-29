package env

import (
	"fmt"
	"os"
)

const (
	OPENAI_API_KEY    = "OPENAI_API_KEY"    // Environment variable name for OpenAI API key
	ANTHROPIC_API_KEY = "ANTHROPIC_API_KEY" // Environment variable name for Anthropic API key
	GEMINI_API_KEY    = "GEMINI_API_KEY"    // Environment variable name for Gemini API key
	YOUTUBE_API_KEY   = "YOUTUBE_API_KEY"   // Environment variable name for Youtube API key
	CHROMA_URL        = "CHROMA_URL"        // Environment variable name for Chroma URL
)

var globalEnvStore = NewEnvStore()

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

type EnvStore struct {
	getEnv func(string) (string, error)
}

func NewEnvStore() *EnvStore {
	return &EnvStore{
		getEnv: func(s string) (string, error) {
			val, ok := os.LookupEnv(s)
			if !ok {
				return "", fmt.Errorf("%s is not set", s)
			}
			return val, nil
		},
	}
}

func (e *EnvStore) Get(key string) (string, error) {
	return e.getEnv(key)
}
