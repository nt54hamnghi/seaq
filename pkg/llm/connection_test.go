package llm

import (
	"maps"
	"slices"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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

type ConnectionSetTestSuite struct {
	suite.Suite
}

func TestConnectionSetTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectionSetTestSuite))
}

func (s *ConnectionSetTestSuite) SetupTest() {
	viper.Reset()
	viper.Set("connections",
		[]map[string]string{
			{
				"provider": "openai",
				"base_url": "https://api.openai.com/v1",
				"env_key":  "OPENAI_API_KEY",
			},
			{
				"provider": "groq",
				"base_url": "https://api.groq.com/openai/v1",
				"env_key":  "GROQ_API_KEY",
			},
		},
	)
}

func (s *ConnectionSetTestSuite) TearDownTest() {
	viper.Reset()
}

func (s *ConnectionSetTestSuite) TestGetConnectionSet_UnmarshalError() {
	viper.Set("connections", "invalid")

	_, err := GetConnectionSet()
	s.Error(err)
}

func (s *ConnectionSetTestSuite) TestGetConnectionSet() {
	r := s.Require()

	cs, err := GetConnectionSet()
	r.NoError(err)

	r.Equal([]Connection{
		{
			Provider: "openai",
			BaseURL:  "https://api.openai.com/v1",
			EnvKey:   "OPENAI_API_KEY",
		},
		{
			Provider: "groq",
			BaseURL:  "https://api.groq.com/openai/v1",
			EnvKey:   "GROQ_API_KEY",
		},
	}, cs.AsSlice())
	r.True(cs.Has("openai"))
	r.False(cs.Has("anthropic"))
}

func (s *ConnectionSetTestSuite) TestConnectionSet_Get() {
	r := s.Require()

	cs, err := GetConnectionSet()
	r.NoError(err)

	testCases := []struct {
		name     string
		provider string
		wantConn Connection
		wantOk   bool
	}{
		{
			name:     "existing provider",
			provider: "groq",
			wantConn: Connection{
				Provider: "groq",
				BaseURL:  "https://api.groq.com/openai/v1",
				EnvKey:   "GROQ_API_KEY",
			},
			wantOk: true,
		},
		{
			name:     "missing provider",
			provider: "anthropic",
			wantConn: Connection{},
			wantOk:   false,
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			conn, ok := cs.Get(tt.provider)
			r.Equal(tt.wantOk, ok)
			r.Equal(tt.wantConn, conn)
		})
	}
}

func (s *ConnectionSetTestSuite) TestConnectionSet_Delete() {
	r := s.Require()

	cs, err := GetConnectionSet()
	r.NoError(err)

	testCases := []struct {
		name              string
		toDelete          string
		wantProvidersLeft []string
	}{
		{
			name:              "existing provider",
			toDelete:          "openai",
			wantProvidersLeft: []string{"groq"},
		},
		{
			name:              "non-existent provider",
			toDelete:          "non-existent",
			wantProvidersLeft: []string{"groq"},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			cs.Delete(tt.toDelete)

			ps := make([]string, 0, len(cs.connections))
			for _, conn := range cs.connections {
				ps = append(ps, conn.Provider)
			}

			r.Equal(tt.wantProvidersLeft, ps)
			r.Equal(tt.wantProvidersLeft, slices.Collect(maps.Keys(cs.index)))
		})
	}
}
