// TESTME:

package llm

import (
	"context"
	"fmt"
	"maps"
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
}

func NewConnection(provider string, baseURL string) Connection {
	return Connection{Provider: provider, BaseURL: baseURL}
}

// GetEnvKey returns the environment variable name that stores the API key
// for the connection. The format is <PROVIDER>_API_KEY.
func (c Connection) GetEnvKey() string {
	return strings.ToUpper(c.Provider) + "_API_KEY"
}

// GetProvider implements the ModelLister interface
// and returns the provider name
func (c Connection) GetProvider() string {
	return c.Provider
}

// List implements the ModelLister interface
// and returns a slice of available model IDs from the provider.
func (c Connection) List(ctx context.Context) ([]string, error) {
	secret, err := env.Get(c.GetEnvKey())
	if err != nil {
		return nil, err
	}

	headers := http.Header{"Authorization": []string{fmt.Sprintf("Bearer %s", secret)}}
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

// ConnectionMap is a map of provider name to Connection.
// This is to provide a fast and ergonomic provider-to-Connection lookup.
// We're trading memory for speed as we don't expect a large number of connections.
// Also, we're paying upfront the cost of constructing the map.
type ConnectionMap map[string]Connection

// GetConnections retrieves the LLM provider connections in the config file
// and constructs a provider-to-Connection map.
func GetConnections() (ConnectionMap, error) {
	connections := []Connection{}
	if err := viper.UnmarshalKey("connections", &connections); err != nil {
		return ConnectionMap{}, err
	}

	connMap := make(ConnectionMap)
	for _, conn := range connections {
		connMap[conn.Provider] = conn
	}

	return connMap, nil
}

// AsSlice returns a slice of all connections in the map.
func (cm ConnectionMap) AsSlice() []Connection {
	return slices.Collect(maps.Values(cm))
}
