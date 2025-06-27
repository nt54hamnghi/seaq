package env

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
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
	JINA_API_KEY       = "JINA_API_KEY"
	FIRECRAWL_API_KEY  = "FIRECRAWL_API_KEY"

	// seaq-specific

	// Deprecated: Use SEAQ_LOG_LEVEL=error to suppress warnings instead.
	SEAQ_SUPPRESS_WARNINGS = "SEAQ_SUPPRESS_WARNINGS" // whether to suppress warnings
	SEAQ_CACHE_DURATION    = "SEAQ_CACHE_DURATION"    // cache duration in seconds
	SEAQ_LOG_LEVEL         = "SEAQ_LOG_LEVEL"         // log level
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

// LogLevel returns the value of the SEAQ_LOG_LEVEL environment variable
// or an error if not set.
func LogLevel() slog.Level {
	lv, err := Get(SEAQ_LOG_LEVEL)
	if err != nil {
		// for backwards compatibility
		if SuppressWarnings() {
			return slog.LevelError
		}
		return slog.LevelInfo
	}
	switch strings.ToLower(strings.TrimSpace(lv)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// SuppressWarnings returns the value of the SEAQ_SUPPRESS_WARNINGS environment variable
// or an error if not set.
//
// Deprecated: SEAQ_SUPPRESS_WARNINGS environment variable is deprecated and replaced by SEAQ_LOG_LEVEL.
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

const defaultCacheDuration = 24 * time.Hour

// CacheDuration returns the value of the SEAQ_CACHE_DURATION environment variable
// or an error if not set.
func CacheDuration() time.Duration {
	val, err := Get(SEAQ_CACHE_DURATION)
	if err != nil {
		return defaultCacheDuration
	}

	dur, err := time.ParseDuration(val)
	if err != nil || dur <= 0 {
		return defaultCacheDuration
	}
	return dur
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

// JinaAPIKey returns the value of the JINA_API_KEY environment variable
// or an error if not set.
func JinaAPIKey() (string, error) {
	return Get(JINA_API_KEY)
}

// FirecrawlAPIKey returns the value of the FIRECRAWL_API_KEY environment variable
// or an error if not set.
func FirecrawlAPIKey() (string, error) {
	return Get(FIRECRAWL_API_KEY)
}
