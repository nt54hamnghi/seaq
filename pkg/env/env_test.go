package env

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EnvSuite struct {
	suite.Suite
	originalValues map[string]string
}

func TestEnvSuite(t *testing.T) {
	suite.Run(t, &EnvSuite{})
}

var supportedEnvs = [13]string{
	OPENAI_API_KEY,
	ANTHROPIC_API_KEY,
	GEMINI_API_KEY,
	YOUTUBE_API_KEY,
	CHROMA_URL,
	X_AUTH_TOKEN,
	X_CSRF_TOKEN,
	UDEMY_ACCESS_TOKEN,
	OLLAMA_HOST,
	JINA_API_KEY,
	FIRECRAWL_API_KEY,
	SEAQ_SUPPRESS_WARNINGS,
	SEAQ_LOG_LEVEL,
}

func (s *EnvSuite) SetupTest() {
	s.originalValues = make(map[string]string)
	for _, key := range supportedEnvs {
		if val, err := Get(key); err == nil {
			s.originalValues[key] = val
		}
		s.originalValues[key] = ""
	}
}

func (s *EnvSuite) TearDownTest() {
	for key, val := range s.originalValues {
		if val == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, val)
		}
	}
}

func (s *EnvSuite) Test_LogLevel() {
	testCases := []struct {
		name     string
		envValue string
		setEnv   bool
		want     slog.Level
	}{
		{
			name:     "debug",
			envValue: "debug",
			setEnv:   true,
			want:     slog.LevelDebug,
		},
		{
			name:     "info",
			envValue: "info",
			setEnv:   true,
			want:     slog.LevelInfo,
		},
		{
			name:     "warn",
			envValue: "warn",
			setEnv:   true,
			want:     slog.LevelWarn,
		},
		{
			name:     "error",
			envValue: "error",
			setEnv:   true,
			want:     slog.LevelError,
		},
		{
			name:     "upperCase",
			envValue: "DEBUG",
			setEnv:   true,
			want:     slog.LevelDebug,
		},
		{
			name:     "mixedCase",
			envValue: "DeBUG",
			setEnv:   true,
			want:     slog.LevelDebug,
		},
		{
			name:     "whitespace",
			envValue: "  debug  ",
			setEnv:   true,
			want:     slog.LevelDebug,
		},
		{
			name:     "invalid",
			envValue: "invalid",
			setEnv:   true,
			want:     slog.LevelInfo,
		},
		{
			name:     "emptyValue",
			envValue: "",
			setEnv:   true,
			want:     slog.LevelInfo,
		},
		{
			name:   "unsetEnv",
			setEnv: false,
			want:   slog.LevelInfo,
		},
	}

	r := s.Require()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.setEnv {
				os.Setenv(SEAQ_LOG_LEVEL, tc.envValue)
			} else {
				os.Unsetenv(SEAQ_LOG_LEVEL)
			}

			// Test
			got := LogLevel()
			r.Equal(tc.want, got)
		})
	}
}

func (s *EnvSuite) Test_LogLevel_BackwardCompatibility() {
	r := s.Require()

	// Ensure SEAQ_LOG_LEVEL is unset so we hit the backward compatibility path
	os.Unsetenv(SEAQ_LOG_LEVEL)

	// Set SEAQ_SUPPRESS_WARNINGS to trigger backward compatibility
	os.Setenv(SEAQ_SUPPRESS_WARNINGS, "true")

	got := LogLevel()
	r.Equal(slog.LevelError, got)
}

func (s *EnvSuite) Test_SuppressWarnings() {
	testCases := []struct {
		name     string
		envValue string
		setEnv   bool
		want     bool
	}{
		{
			name:     "true",
			envValue: "true",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "1",
			envValue: "1",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "yes",
			envValue: "yes",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "y",
			envValue: "y",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "on",
			envValue: "on",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "upperCase",
			envValue: "TRUE",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "mixedCase",
			envValue: "YeS",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "whitespace",
			envValue: "  true  ",
			setEnv:   true,
			want:     true,
		},
		{
			name:     "false",
			envValue: "false",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "0",
			envValue: "0",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "invalid",
			envValue: "invalid",
			setEnv:   true,
			want:     false,
		},
		{
			name:     "empty",
			envValue: "",
			setEnv:   true,
			want:     false,
		},
		{
			name:   "unset",
			setEnv: false,
			want:   false,
		},
	}

	r := s.Require()

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.setEnv {
				os.Setenv(SEAQ_SUPPRESS_WARNINGS, tc.envValue)
			} else {
				os.Unsetenv(SEAQ_SUPPRESS_WARNINGS)
			}

			got := SuppressWarnings()
			r.Equal(tc.want, got)
		})
	}
}

func (s *EnvSuite) Test_OllamaHost() {
	r := s.Require()

	// Test with env var set
	os.Setenv(OLLAMA_HOST, "http://custom-host:8080")
	got := OllamaHost()
	r.Equal("http://custom-host:8080", got)

	// Test with env var unset (should return default)
	os.Unsetenv(OLLAMA_HOST)
	got = OllamaHost()
	r.Equal("http://localhost:11434", got)
}
