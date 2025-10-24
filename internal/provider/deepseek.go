package provider

import "net/http"

// DeepSeek provider implementation
type DeepSeek struct{}

func (d *DeepSeek) GetEndpoint() string {
	return "https://api.deepseek.com/v1/chat/completions"
}

func (d *DeepSeek) SetAuthHeaders(req *http.Request, apiKey string) {
	req.Header.Set("Authorization", "Bearer "+apiKey)
}

func (d *DeepSeek) GetName() string {
	return "deepseek"
}
