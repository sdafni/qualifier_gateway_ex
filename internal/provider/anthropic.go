package provider

import "net/http"

// Anthropic provider implementation
type Anthropic struct{}

func (a *Anthropic) GetEndpoint() string {
	return "https://api.anthropic.com/v1/messages"
}

func (a *Anthropic) SetAuthHeaders(req *http.Request, apiKey string) {
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

func (a *Anthropic) GetName() string {
	return "anthropic"
}
