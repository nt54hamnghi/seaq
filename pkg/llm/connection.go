package llm

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/spf13/viper"
)

// identRegex defines the regex for valid identifiers
var identRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// IsIdent checks if a string is a valid identifier
func IsIdent(s string) bool {
	return identRegex.MatchString(s)
}

// Connection represents a single API connection configuration
type Connection struct {
	Provider string `mapstructure:"provider" yaml:"provider"`
	BaseURL  string `mapstructure:"base_url" yaml:"base_url"`
	EnvKey   string `mapstructure:"env_key" yaml:"env_key"`
}

func NewConnection(provider string, baseURL string, envKey string) Connection {
	if envKey == "" {
		envKey = strings.ToUpper(provider) + "_API_KEY"
	}
	return Connection{provider, baseURL, envKey}
}

// GetProvider implements the ModelLister interface
// and returns the provider name
func (c Connection) GetProvider() string {
	return c.Provider
}

// List implements the ModelLister interface
// and returns a slice of available model IDs from the provider.
func (c Connection) List(ctx context.Context) ([]string, error) {
	secret, err := env.Get(c.EnvKey)
	if err != nil {
		return nil, err
	}

	headers := http.Header{"Authorization": []string{"Bearer " + secret}}
	res, err := reqx.GetAs[listModelsResponse](ctx, c.BaseURL+"/models", headers)
	if err != nil {
		return nil, fmt.Errorf("fetch models: %w", err)
	}

	models := make([]string, len(res.Data))
	for i, model := range res.Data {
		models[i] = model.ID
	}
	return models, nil
}

type listModelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		ID      string `json:"id"`
		Created int    `json:"created"`
		Object  string `json:"object"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// ConnectionSet stores configured connections and a provider index for efficient lookup
type ConnectionSet struct {
	// connections keeps the original list loaded from config.
	connections []Connection
	// index maps provider names to connection entries for O(1) lookup.
	index map[string]*Connection
}

// GetConnectionSet loads configured connections from viper
// and builds the collection set.
func GetConnectionSet() (ConnectionSet, error) {
	connections := []Connection{}
	if err := viper.UnmarshalKey("connections", &connections); err != nil {
		return ConnectionSet{}, err
	}

	index := make(map[string]*Connection)
	for i := range connections {
		c := connections[i]
		index[c.Provider] = &c
	}

	return ConnectionSet{
		connections,
		index,
	}, nil
}

// AsSlice returns the current connections as a slice.
func (cs ConnectionSet) AsSlice() []Connection {
	return cs.connections
}

// Has reports whether a provider exists in the collection set.
func (cs ConnectionSet) Has(provider string) bool {
	_, ok := cs.index[provider]
	return ok
}

// Get returns a connection by provider and whether it exists.
func (cs ConnectionSet) Get(provider string) (Connection, bool) {
	c, ok := cs.index[provider]
	if !ok || c == nil {
		return Connection{}, false
	}
	return *c, true
}

// Delete removes a provider from the collection set
func (cs *ConnectionSet) Delete(provider string) {
	cs.connections = slices.DeleteFunc(cs.connections, func(c Connection) bool { return c.Provider == provider })
	delete(cs.index, provider)
}
