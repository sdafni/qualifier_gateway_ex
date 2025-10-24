package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Gateway endpoint
	chatCompletionsPath = "/chat/completions"

	// Default port
	defaultPort = "8080"

	// Default keys file
	defaultKeysFile = "keys.json"

	// Logging
	logsDir     = "logs"
	logFilename = "llm_interactions.jsonl"

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

// LogEntry represents a logged LLM interaction
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	VirtualKey string                 `json:"virtual_key"`
	Provider   string                 `json:"provider"`
	Method     string                 `json:"method"`
	Status     int                    `json:"status"`
	DurationMs int64                  `json:"duration_ms"`
	Request    map[string]interface{} `json:"request"`
	Response   map[string]interface{} `json:"response"`
}

type Gateway struct {
	config  *Config
	logFile *os.File
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

func InitLogging() (*os.File, error) {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Open log file in append mode
	logPath := fmt.Sprintf("%s/%s", logsDir, logFilename)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return logFile, nil
}

func NewGateway(config *Config, logFile *os.File) *Gateway {
	return &Gateway{
		config:  config,
		logFile: logFile,
	}
}

func (g *Gateway) logInteraction(entry LogEntry) {
	// Log to console (pretty-printed)
	consoleJSON, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		log.Printf("Error marshaling log entry for console: %v", err)
	} else {
		log.Printf("LLM Interaction Log:\n%s", string(consoleJSON))
	}

	// Log to file (single line JSON)
	fileJSON, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Error marshaling log entry for file: %v", err)
		return
	}

	if _, err := g.logFile.Write(append(fileJSON, '\n')); err != nil {
		log.Printf("Error writing to log file: %v", err)
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Start timing
	startTime := time.Now()

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

	// Read and buffer the request body
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		log.Printf("Error reading request body: %v", err)
		return
	}
	defer r.Body.Close()

	// Parse request JSON for logging
	var requestJSON map[string]interface{}
	if err := json.Unmarshal(requestBody, &requestJSON); err != nil {
		log.Printf("Warning: Failed to parse request JSON for logging: %v", err)
		requestJSON = map[string]interface{}{"raw": string(requestBody)}
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

	// Create a new request with buffered body
	proxyReq, err := http.NewRequest(r.Method, targetURL, bytes.NewReader(requestBody))
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

	// Read and buffer the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusBadGateway)
		log.Printf("Error reading response body: %v", err)
		return
	}

	// Parse response JSON for logging
	var responseJSON map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseJSON); err != nil {
		log.Printf("Warning: Failed to parse response JSON for logging: %v", err)
		responseJSON = map[string]interface{}{"raw": string(responseBody)}
	}

	// Calculate duration
	duration := time.Since(startTime)

	// Create log entry
	logEntry := LogEntry{
		Timestamp:  startTime.Format(time.RFC3339),
		VirtualKey: virtualKey,
		Provider:   keyConfig.Provider,
		Method:     r.Method,
		Status:     resp.StatusCode,
		DurationMs: duration.Milliseconds(),
		Request:    requestJSON,
		Response:   responseJSON,
	}

	// Log the interaction
	g.logInteraction(logEntry)

	// Copy response headers to client
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Write buffered response body to client
	if _, err := w.Write(responseBody); err != nil {
		log.Printf("Error writing response body to client: %v", err)
	}

	log.Printf("Request completed with status: %d (provider: %s, duration: %dms)", resp.StatusCode, keyConfig.Provider, duration.Milliseconds())
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

	// Initialize logging
	logFile, err := InitLogging()
	if err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}
	defer logFile.Close()

	// Create gateway
	gateway := NewGateway(config, logFile)

	log.Printf("Starting LLM Gateway Router on port %s", port)
	log.Printf("Loaded configuration from: %s", keysFile)
	log.Printf("Logging to: %s/%s", logsDir, logFilename)
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
