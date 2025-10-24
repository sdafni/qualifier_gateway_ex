package provider

import "net/http"

// OpenAI provider implementation
type OpenAI struct{}

func (o *OpenAI) GetEndpoint() string {
	return "https://api.openai.com/v1/chat/completions"
}

func (o *OpenAI) SetAuthHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
}

func (o *OpenAI) GetName() string {
	return "openai"
}
