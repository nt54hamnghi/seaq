package llm

import (
	"errors"
	"fmt"
	"iter"
	"log"
	"maps"
	"strings"
	"sync"

	"github.com/nt54hamnghi/seaq/pkg/util/set"
)

type ModelRegistry map[string]set.Set[string]

// ModelRetriever is a function that returns a list of model names
type ModelRetriever func() ([]string, error)

var (
	ErrProviderNameEmpty     = errors.New("provider name cannot be empty")
	ErrModelsListEmpty       = errors.New("models list cannot be empty")
	ErrProviderAlreadyExists = errors.New("provider already exists")
)

var initOnce sync.Once

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

func initRegistry() {
	registrars := []struct {
		provider  string
		retriever ModelRetriever
	}{
		{"ollama", listOllamaModels},
	}

	initOnce.Do(func() {
		for _, r := range registrars {
			if err := Registry.RegisterWith(r.provider, r.retriever); err != nil {
				log.Printf("[WARN] %s registration: %v\n\n", r.provider, err)
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

// RegisterWith adds a new provider and its models to the registry using a function.
// The function fn should return a slice of model names.
//
// Returns error if:
//   - provider name is empty
//   - models list is empty
//   - provider already exists
//   - fn fails to retrieve the models
func (r ModelRegistry) RegisterWith(provider string, retriever ModelRetriever) error {
	models, err := retriever()
	if err != nil {
		return err
	}
	return r.Register(provider, models)
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

// RegisterFunc adds a new provider and its models to the default registry using a function.
// See ModelRegistry.RegisterFunc for details.
func RegisterFunc(provider string, fn func() ([]string, error)) error {
	initRegistry()
	return Registry.RegisterWith(provider, fn)
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
