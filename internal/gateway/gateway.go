package gateway

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"gateway/internal/logger"
	"gateway/internal/provider"
	"gateway/internal/virtualkey"
)

const chatCompletionsPath = "/chat/completions"

// Gateway handles incoming requests and routes them to appropriate providers
type Gateway struct {
	vkService        *virtualkey.Service
	providerRegistry *provider.Registry
	logger           *logger.Logger
}

// New creates a new Gateway instance
func New(vkService *virtualkey.Service, providerRegistry *provider.Registry, log *logger.Logger) *Gateway {
	return &Gateway{
		vkService:        vkService,
		providerRegistry: providerRegistry,
		logger:           log,
	}
}

// ServeHTTP implements the http.Handler interface
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

	// Validate virtual key
	virtualKey, err := g.vkService.ValidateRequest(r.Header.Get("Authorization"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		log.Printf("Request rejected: %v", err)
		return
	}

	// Get key configuration
	keyConfig, exists := g.vkService.GetKeyConfig(virtualKey)
	if !exists {
		http.Error(w, "Invalid virtual key", http.StatusUnauthorized)
		log.Printf("Request rejected: virtual key not found: %s", virtualKey)
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
	requestJSON := logger.ParseJSONBody(requestBody, r.Header.Get("Content-Encoding"))

	// Get provider
	prov, err := g.providerRegistry.Get(keyConfig.Provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error getting provider: %v", err)
		return
	}

	log.Printf("Routing request to %s provider (virtual key: %s)", prov.GetName(), virtualKey)

	// Create proxy request
	proxyReq, err := http.NewRequest(r.Method, prov.GetEndpoint(), bytes.NewReader(requestBody))
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

	// Set provider-specific authentication headers
	prov.SetAuthHeaders(proxyReq, keyConfig.APIKey)

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

	// Parse response body for logging (handles decompression internally)
	responseJSON := logger.ParseJSONBody(responseBody, resp.Header.Get("Content-Encoding"))

	// Calculate duration
	duration := time.Since(startTime)

	// Create and log interaction
	logEntry := logger.LogEntry{
		Timestamp:  startTime.Format(time.RFC3339),
		VirtualKey: virtualKey,
		Provider:   prov.GetName(),
		Method:     r.Method,
		Status:     resp.StatusCode,
		DurationMs: duration.Milliseconds(),
		Request:    requestJSON,
		Response:   responseJSON,
	}
	g.logger.LogInteraction(logEntry)

	// Copy response headers to client
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set status code and write response
	w.WriteHeader(resp.StatusCode)
	if _, err := w.Write(responseBody); err != nil {
		log.Printf("Error writing response body to client: %v", err)
	}

	log.Printf("Request completed with status: %d (provider: %s, duration: %dms)",
		resp.StatusCode, prov.GetName(), duration.Milliseconds())
}
