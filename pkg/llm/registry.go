package llm

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"strings"
	"sync"

	"github.com/nt54hamnghi/seaq/pkg/util/log"
	"github.com/nt54hamnghi/seaq/pkg/util/set"
)

type ModelRegistry map[string]set.Set[string]

// ModelLister defines an interface for services that can list available models.
type ModelLister interface {
	// GetProvider returns the name of the model provider (e.g., "openai", "anthropic")
	GetProvider() string

	// List returns a slice of available model IDs from the provider.
	// Returns an error if the listing operation fails.
	List(context.Context) ([]string, error)
}

// SimpleModelLister provides a simple implementation of the ModelLister interface
// that allows creating a model lister with a provider name and a listing function.
type SimpleModelLister struct {
	ProviderName string
	Lister       func(context.Context) ([]string, error)
}

func (l SimpleModelLister) GetProvider() string {
	return l.ProviderName
}

func (l SimpleModelLister) List(ctx context.Context) ([]string, error) {
	return l.Lister(ctx)
}

var (
	ErrProviderNameEmpty     = errors.New("provider name cannot be empty")
	ErrModelsListEmpty       = errors.New("models list cannot be empty")
	ErrProviderAlreadyExists = errors.New("provider already exists")
)

// Default registry of models
var Registry = ModelRegistry{
	"openai": {
		// O1:            {},
		// O1Mini:        {},
		// O1Preview:     {},
		GPT4o:         {},
		GPT4oMini:     {},
		GPT4:          {},
		GPT4Turbo:     {},
		ChatGPT4o:     {},
		GPT3Dot5Turbo: {},
	},
	"anthropic": {
		Claude3Dot5Sonnet: {},
		Claude3Dot5Haiku:  {},
		Claude3Opus:       {},
		Claude3Sonnet:     {},
		Claude3Haiku:      {},
	},
	"google": {
		Gemini2Dot0FlashExp: {},
		Gemini1Dot5Flash:    {},
		Gemini1Dot5Flash8B:  {},
		Gemini1Dot5Pro:      {},
	},
}

var (
	initOnce sync.Once
	connMap  ConnectionMap
)

func initRegistry() {
	initOnce.Do(func() {
		var err error

		ctx := context.Background()

		listers := []ModelLister{ollamaLister}
		connMap, err = GetConnections()
		if err != nil {
			log.Warn("failed to load connections", "error", err)
		}
		for conn := range maps.Values(connMap) {
			listers = append(listers, conn)
		}

		for _, l := range listers {
			if err := Registry.RegisterWith(ctx, l); err != nil {
				log.Warn("failed to register models", "provider", l.GetProvider(), "error", err)
			}
		}
	})
}

// normalize returns a string that is trimmed of whitespace and converted to lowercase.
func normalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

func toModelID(provider, model string) string {
	return fmt.Sprintf("%s/%s", provider, model)
}

// LookupModel returns the provider and model name for a given model identifier.
// Model identifier must follow the format "provider/model" (e.g., "openai/gpt-4").
// Both provider and model names are case-insensitive and trimmed of whitespace.
//
// Returns:
//   - provider: The provider name in lowercase if found, empty string if not found
//   - model: The model name in lowercase if found, empty string if not found
//   - ok: True if the model was found, false otherwise
//
// The method returns empty strings and false if:
//   - The input id is empty
//   - The input doesn't contain the required "/" separator
//   - The provider/model format is used but either part is empty after trimming
//   - The specified provider doesn't exist
//   - The specified model doesn't exist for the given provider
func (r ModelRegistry) LookupModel(id string) (provider, model string, ok bool) {
	if id == "" {
		return "", "", false
	}

	// If we have a provider/model format, look up that specific combination
	if provider, model, hasSep := strings.Cut(id, "/"); hasSep {
		provider, model = normalize(provider), normalize(model)

		if provider == "" || model == "" {
			return "", "", false
		}

		models, exists := r[provider]
		if !exists || !models.Contains(model) {
			return "", "", false
		}
		return provider, model, true
	}

	return "", "", false
}

// HasModel checks if a model is supported by any provider in the registry.
func (r ModelRegistry) HasModel(name string) bool {
	_, _, ok := r.LookupModel(name)
	return ok
}

// Register adds a new provider and its associated models to the registry.
// Both provider and model names are normalized (trimmed of whitespace and converted to lowercase).
//
// Returns error if:
//   - provider name is empty after normalization
//   - models list is empty
//   - provider already exists
func (r ModelRegistry) Register(provider string, models []string) error {
	if len(models) == 0 {
		return ErrModelsListEmpty
	}

	provider = normalize(provider)
	if provider == "" {
		return ErrProviderNameEmpty
	}

	if _, ok := r[provider]; ok {
		return ErrProviderAlreadyExists
	}

	// Normalize all model names
	normalizedModels := make([]string, len(models))
	for i, model := range models {
		normalizedModels[i] = normalize(model)
	}

	r[provider] = set.New(normalizedModels...)
	return nil
}

// RegisterWith adds a new provider and its models to the registry using a ModelLister.
//
// Returns error if:
//   - provider name is empty after normalization
//   - models list is empty
//   - provider already exists in registry
//   - ModelLister.List fails to retrieve the models
func (r ModelRegistry) RegisterWith(ctx context.Context, l ModelLister) error {
	models, err := l.List(ctx)
	if err != nil {
		return err
	}
	return r.Register(l.GetProvider(), models)
}

// Iter returns an iterator that yields provider-model pairs from the registry.
// The iterator yields each model name along with its provider.
// The order of iteration is non-deterministic.
// Example:
//
//	for provider, model := range registry.Iter() {
//	    fmt.Printf("Provider: %s, Model: %s\n", provider, model)
//	}
func (r ModelRegistry) Iter() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for provider, models := range r {
			for m := range models.Iter() {
				if !yield(provider, m) {
					return
				}
			}
		}
	}
}

// Providers returns an iterator over all providers in the registry.
// The iterator yields each provider name as a string.
// The order of providers is non-deterministic
func (r ModelRegistry) Providers() iter.Seq[string] {
	return maps.Keys(r)
}

// ModelsByProvider returns an iterator over all models for a given provider.
// Each iteration yields a model identifier (in the format "provider/model") as a string.
// The order is non-deterministic
func (r ModelRegistry) ModelsByProvider(provider string) iter.Seq[string] {
	return func(yield func(string) bool) {
		models, ok := r[provider]
		if !ok {
			return
		}
		for m := range models.Iter() {
			id := toModelID(provider, m)
			if !yield(id) {
				return
			}
		}
	}
}

// Models returns an iterator over all models in the registry.
// Each iteration yields a model identifier (in the format "provider/model") as a string.
// The order is non-deterministic
func (r ModelRegistry) Models() iter.Seq[string] {
	return func(yield func(string) bool) {
		for p, m := range r.Iter() {
			id := toModelID(p, m)
			if !yield(id) {
				return
			}
		}
	}
}

// Top-level functions that operate on the default Registry

// LookupModel looks up a model in the default registry.
// See ModelRegistry.LookupModel for details.
func LookupModel(id string) (provider, model string, ok bool) {
	initRegistry()
	return Registry.LookupModel(id)
}

// HasModel checks if a model is supported by any provider in the default registry.
// See ModelRegistry.HasModel for details.
func HasModel(name string) bool {
	initRegistry()
	return Registry.HasModel(name)
}

// Register adds a new provider and its models to the default registry.
// See ModelRegistry.Register for details.
func Register(provider string, models []string) error {
	initRegistry()
	return Registry.Register(provider, models)
}

// RegisterWith adds a new provider and its models to the default registry using a ModelLister.
// See ModelRegistry.RegisterWith for details.
func RegisterWith(ctx context.Context, l ModelLister) error {
	initRegistry()
	return Registry.RegisterWith(ctx, l)
}

// Providers returns an iterator over all providers in the default registry.
// See ModelRegistry.Providers for details.
func Providers() iter.Seq[string] {
	initRegistry()
	return Registry.Providers()
}

// ModelsByProvider returns an iterator over all models for a given provider in the default registry.
// See ModelRegistry.ModelsByProvider for details.
func ModelsByProvider(provider string) iter.Seq[string] {
	initRegistry()
	return Registry.ModelsByProvider(provider)
}

// Models returns an iterator over all models in the default registry.
// See ModelRegistry.Models for details.
func Models() iter.Seq[string] {
	initRegistry()
	return Registry.Models()
}
