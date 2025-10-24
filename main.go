package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	// Gateway endpoint
	chatCompletionsPath = "/chat/completions"

	// Default port
	defaultPort = "8080"

	// Default keys file
	defaultKeysFile = "keys.json"

	// Provider URLs
	openAIURL    = "https://api.openai.com/v1/chat/completions"
	anthropicURL = "https://api.anthropic.com/v1/messages"
	deepseekURL  = "https://api.deepseek.com/v1/chat/completions"
)

// KeyConfig represents a single virtual key configuration
type KeyConfig struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

// Config represents the keys.json structure
type Config struct {
	VirtualKeys map[string]KeyConfig `json:"virtual_keys"`
}

type Gateway struct {
	config *Config
}

func LoadConfig(filepath string) (*Config, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return &config, nil
}

func NewGateway(config *Config) *Gateway {
	return &Gateway{
		config: config,
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only handle the chat completions endpoint
	if r.URL.Path != chatCompletionsPath {
		http.NotFound(w, r)
		return
	}

	// Only handle POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract virtual key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		log.Println("Request rejected: missing Authorization header")
		return
	}

	// Extract bearer token
	virtualKey := strings.TrimPrefix(authHeader, "Bearer ")
	if virtualKey == authHeader {
		http.Error(w, "Invalid Authorization header format. Expected 'Bearer <virtual-key>'", http.StatusUnauthorized)
		log.Println("Request rejected: invalid Authorization header format")
		return
	}

	// Look up virtual key in config
	keyConfig, exists := g.config.VirtualKeys[virtualKey]
	if !exists {
		http.Error(w, "Invalid virtual key", http.StatusUnauthorized)
		log.Printf("Request rejected: invalid virtual key: %s", virtualKey)
		return
	}

	// Determine target URL based on provider
	var targetURL string
	switch strings.ToLower(keyConfig.Provider) {
	case "openai":
		targetURL = openAIURL
	case "anthropic":
		targetURL = anthropicURL
	case "deepseek":
		targetURL = deepseekURL
	default:
		http.Error(w, "Unsupported provider", http.StatusInternalServerError)
		log.Printf("Unsupported provider: %s", keyConfig.Provider)
		return
	}

	log.Printf("Routing request to %s provider (virtual key: %s)", keyConfig.Provider, virtualKey)

	// Create a new request to forward
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		log.Printf("Error creating proxy request: %v", err)
		return
	}

	// Copy headers from original request (excluding Authorization)
	for name, values := range r.Header {
		if name != "Authorization" {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}
	}

	// Set the real API key based on provider
	switch strings.ToLower(keyConfig.Provider) {
	case "openai", "deepseek":
		// OpenAI and DeepSeek use Bearer token authentication
		proxyReq.Header.Set("Authorization", "Bearer "+keyConfig.APIKey)
	case "anthropic":
		// Anthropic uses x-api-key header
		proxyReq.Header.Set("x-api-key", keyConfig.APIKey)
		proxyReq.Header.Set("anthropic-version", "2023-06-01")
	}

	// Forward the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		log.Printf("Error forwarding request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("Error copying response body: %v", err)
	}

	log.Printf("Request completed with status: %d (provider: %s)", resp.StatusCode, keyConfig.Provider)
}

func main() {
	// Get configuration from environment variables or use defaults
	keysFile := os.Getenv("KEYS_FILE")
	if keysFile == "" {
		keysFile = defaultKeysFile
	}

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = defaultPort
	}

	// Load configuration
	config, err := LoadConfig(keysFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create gateway
	gateway := NewGateway(config)

	log.Printf("Starting LLM Gateway Router on port %s", port)
	log.Printf("Loaded configuration from: %s", keysFile)
	log.Printf("Endpoint: POST %s", chatCompletionsPath)
	log.Printf("Configured virtual keys: %d", len(config.VirtualKeys))
	log.Println("Virtual key mappings:")
	for vk, kc := range config.VirtualKeys {
		log.Printf("  %s -> %s provider", vk, kc.Provider)
	}

	if err := http.ListenAndServe(":"+port, gateway); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
