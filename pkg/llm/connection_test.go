package llm

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	testCases := []struct {
		name     string
		provider string
		baseURL  string
		envKey   string
		want     Connection
	}{
		{
			name:     "default env key if empty",
			provider: "openai",
			baseURL:  "https://api.openai.com/v1",
			envKey:   "",
			want: Connection{
				Provider: "openai",
				BaseURL:  "https://api.openai.com/v1",
				EnvKey:   "OPENAI_API_KEY",
			},
		},
		{
			name:     "use provided env key",
			provider: "groq",
			baseURL:  "https://api.groq.com/openai/v1",
			envKey:   "GROQ_SECRET",
			want: Connection{
				Provider: "groq",
				BaseURL:  "https://api.groq.com/openai/v1",
				EnvKey:   "GROQ_SECRET",
			},
		},
	}

	r := require.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := NewConnection(tt.provider, tt.baseURL, tt.envKey)
			r.Equal(tt.want, got)
		})
	}
}
