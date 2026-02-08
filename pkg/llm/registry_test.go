package llm

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/nt54hamnghi/seaq/pkg/util/set"
	"github.com/stretchr/testify/require"
)

func TestLookupModel(t *testing.T) {
	registry := ModelRegistry{
		"openai":    set.New("gpt-4o"),
		"anthropic": set.New("claude-3-5-sonnet"),
		"google":    set.New("gemini-2-0-flash"),
	}

	testCases := []struct {
		name         string
		registry     ModelRegistry
		id           string
		wantProvider string
		wantModel    string
		wantFound    bool
	}{
		{
			name:         "success",
			registry:     registry,
			id:           "openai/gpt-4o",
			wantProvider: "openai",
			wantModel:    "gpt-4o",
			wantFound:    true,
		},
		{
			name:         "non-existent provider",
			registry:     registry,
			id:           "non-existent/gpt-4o",
			wantProvider: "",
			wantModel:    "",
			wantFound:    false,
		},
		{
			name:         "non-existent model",
			registry:     registry,
			id:           "openai/non-existent",
			wantProvider: "",
			wantModel:    "",
			wantFound:    false,
		},
		{
			name:         "leading slash",
			registry:     registry,
			id:           "/gpt-4o",
			wantProvider: "",
			wantModel:    "",
			wantFound:    false,
		},
		{
			name:         "trailing slash",
			registry:     registry,
			id:           "openai/",
			wantProvider: "",
			wantModel:    "",
			wantFound:    false,
		},

		{
			name:         "empty registry",
			registry:     ModelRegistry{},
			id:           "openai/gpt-4o",
			wantProvider: "",
			wantModel:    "",
			wantFound:    false,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			provider, model, found := tt.registry.LookupModel(tt.id)
			r.Equal(tt.wantProvider, provider)
			r.Equal(tt.wantModel, model)
			r.Equal(tt.wantFound, found)
		})
	}
}

func TestRegister(t *testing.T) {
	registry := ModelRegistry{
		"openai":    set.New("gpt-4o"),
		"anthropic": set.New("claude-3-5-sonnet"),
		"google":    set.New("gemini-2-0-flash"),
	}

	testCases := []struct {
		name     string
		registry ModelRegistry
		provider string
		models   []string
		wantErr  error
	}{
		{
			name:     "new provider",
			registry: registry,
			provider: "TestProvider",
			models:   []string{"model-1", "model-2"},
			wantErr:  nil,
		},
		{
			name:     "empty provider",
			registry: registry,
			provider: "",
			models:   []string{"model-1", "model-2"},
			wantErr:  ErrProviderNameEmpty,
		},
		{
			name:     "empty models",
			registry: registry,
			provider: "EmptyProvider",
			models:   []string{},
			wantErr:  ErrModelsListEmpty,
		},
		{
			name:     "existing provider",
			registry: registry,
			provider: "openai",
			models:   []string{"model-1", "model-2"},
			wantErr:  ErrProviderAlreadyExists,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			err := tt.registry.Register(tt.provider, tt.models)

			if tt.wantErr != nil {
				r.EqualError(tt.wantErr, err.Error())
				return
			}

			models, ok := tt.registry[clean(tt.provider)]
			r.True(ok)
			r.Equal(set.New(tt.models...), models)
		})
	}
}

func TestRegisterWith(t *testing.T) {
	registry := ModelRegistry{
		"openai":    set.New("gpt-4o"),
		"anthropic": set.New("claude-3-5-sonnet"),
		"google":    set.New("gemini-2-0-flash"),
	}

	testCases := []struct {
		name     string
		registry ModelRegistry
		provider string
		fn       func(context.Context) ([]string, error)
		wantErr  error
	}{
		{
			name:     "new provider",
			registry: registry,
			provider: "TestProvider",
			fn: func(_ context.Context) ([]string, error) {
				return []string{"model-1", "model-2"}, nil
			},
			wantErr: nil,
		},
		{
			name:     "empty provider",
			registry: registry,
			provider: "",
			fn: func(_ context.Context) ([]string, error) {
				return []string{"model-1", "model-2"}, nil
			},
			wantErr: ErrProviderNameEmpty,
		},
		{
			name:     "empty models from func",
			registry: registry,
			provider: "EmptyProvider",
			fn: func(_ context.Context) ([]string, error) {
				return []string{}, nil
			},
			wantErr: ErrModelsListEmpty,
		},
		{
			name:     "existing provider",
			registry: registry,
			provider: "openai",
			fn: func(_ context.Context) ([]string, error) {
				return []string{"model-1", "model-2"}, nil
			},
			wantErr: ErrProviderAlreadyExists,
		},
		{
			name:     "function error",
			registry: registry,
			provider: "ErrorProvider",
			fn: func(_ context.Context) ([]string, error) {
				return nil, errors.New("failed to fetch models")
			},
			wantErr: errors.New("failed to fetch models"),
		},
	}

	r := require.New(t)
	ctx := context.Background()

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			lister := SimpleModelLister{
				ProviderName: tt.provider,
				Lister:       tt.fn,
			}
			err := tt.registry.RegisterWith(ctx, lister)

			if tt.wantErr != nil {
				r.EqualError(tt.wantErr, err.Error())
				return
			}

			models, ok := tt.registry[clean(tt.provider)]
			r.True(ok)

			// Get models from the function to verify
			expectedModels, _ := tt.fn(ctx)
			r.Equal(set.New(expectedModels...), models)
		})
	}
}

func TestModelsByProvider(t *testing.T) {
	testCases := []struct {
		name     string
		registry ModelRegistry
		provider string
		want     []string
	}{
		{
			name:     "empty registry",
			registry: ModelRegistry{},
			provider: "openai",
			want:     []string(nil),
		},
		{
			name: "existing provider",
			registry: ModelRegistry{
				"openai": set.New("gpt-3.5", "gpt-4"),
			},
			provider: "openai",
			want:     []string{"openai/gpt-3.5", "openai/gpt-4"},
		},
		{
			name: "non-existent provider",
			registry: ModelRegistry{
				"openai": set.New("gpt-3.5", "gpt-4"),
			},
			provider: "non-existent",
			want:     []string(nil),
		},
		{
			name: "empty provider name",
			registry: ModelRegistry{
				"openai": set.New("gpt-3.5", "gpt-4"),
			},
			provider: "",
			want:     []string(nil),
		},
		{
			name: "multiple providers",
			registry: ModelRegistry{
				"openai":    set.New("gpt-3.5", "gpt-4"),
				"anthropic": set.New("claude-3"),
			},
			provider: "openai",
			want:     []string{"openai/gpt-3.5", "openai/gpt-4"},
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := slices.Collect(tt.registry.ModelsByProvider(tt.provider))
			slices.Sort(got)

			r.Equal(tt.want, got)
		})
	}
}

func TestModels(t *testing.T) {
	testCases := []struct {
		name     string
		registry ModelRegistry
		want     []string
	}{
		{
			name:     "empty registry",
			registry: ModelRegistry{},
			want:     []string(nil),
		},
		{
			name: "single provider",
			registry: ModelRegistry{
				"openai": set.New("gpt-3.5", "gpt-4"),
			},
			want: []string{"openai/gpt-3.5", "openai/gpt-4"},
		},
		{
			name: "multiple providers",
			registry: ModelRegistry{
				"openai":    set.New("gpt-4"),
				"anthropic": set.New("claude-3"),
				"google":    set.New("gemini-pro"),
			},
			want: []string{"anthropic/claude-3", "google/gemini-pro", "openai/gpt-4"},
		},
		{
			name: "overlapping model names",
			registry: ModelRegistry{
				"openai":     set.New("gpt-4"),
				"openrouter": set.New("gpt-4"),
			},
			want: []string{"openai/gpt-4", "openrouter/gpt-4"},
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := slices.Collect(tt.registry.Models())
			slices.Sort(got)

			r.Equal(tt.want, got)
		})
	}
}

func TestIter(t *testing.T) {
	testCases := []struct {
		name     string
		registry ModelRegistry
		want     []string
	}{
		{
			name:     "empty registry",
			registry: ModelRegistry{},
			want:     []string(nil),
		},
		{
			name: "single provider single model",
			registry: ModelRegistry{
				"openai": set.New("gpt-4"),
			},
			want: []string{"openai/gpt-4"},
		},
		{
			name: "single provider multiple models",
			registry: ModelRegistry{
				"openai": set.New("gpt-3.5", "gpt-4"),
			},
			want: []string{"openai/gpt-3.5", "openai/gpt-4"},
		},
		{
			name: "multiple providers",
			registry: ModelRegistry{
				"openai":    set.New("gpt-3.5", "gpt-4"),
				"anthropic": set.New("claude-2", "claude-3"),
				"google":    set.New("gemini-pro"),
			},
			want: []string{
				"anthropic/claude-2",
				"anthropic/claude-3",
				"google/gemini-pro",
				"openai/gpt-3.5",
				"openai/gpt-4",
			},
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			var got []string
			for p, m := range tt.registry.Iter() {
				got = append(got, fmt.Sprintf("%s/%s", p, m))
			}
			slices.Sort(got)
			r.Equal(tt.want, got)
		})
	}
}
