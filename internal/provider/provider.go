package provider

import (
	"fmt"
	"net/http"
	"strings"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// GetEndpoint returns the API endpoint URL
	GetEndpoint() string

	// SetAuthHeaders sets the authentication headers on the request
	SetAuthHeaders(req *http.Request, apiKey string)

	// GetName returns the provider name
	GetName() string
}

// Registry manages provider instances
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: map[string]Provider{
			"openai":    &OpenAI{},
			"anthropic": &Anthropic{},
			"deepseek":  &DeepSeek{},
		},
	}
}

// Get returns a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	provider, exists := r.providers[strings.ToLower(name)]
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", name)
	}
	return provider, nil
}
