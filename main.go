package main

import (
	"log"
	"net/http"
	"os"

	"gateway/internal/config"
	"gateway/internal/gateway"
	"gateway/internal/logger"
	"gateway/internal/provider"
	"gateway/internal/virtualkey"
)

const (
	defaultPort     = "8080"
	defaultKeysFile = "keys.json"
)

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
	cfg, err := config.Load(keysFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	lgr, err := logger.New()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer lgr.Close()

	// Initialize virtual key service
	vkService := virtualkey.New(cfg)

	// Initialize provider registry
	providerRegistry := provider.NewRegistry()

	// Create gateway
	gw := gateway.New(vkService, providerRegistry, lgr)

	// Log startup information
	log.Printf("Starting LLM Gateway Router on port %s", port)
	log.Printf("Loaded configuration from: %s", keysFile)
	log.Printf("Endpoint: POST /chat/completions")
	log.Printf("Configured virtual keys: %d", len(cfg.VirtualKeys))
	log.Println("Virtual key mappings:")
	for vk, kc := range cfg.VirtualKeys {
		log.Printf("  %s -> %s provider", vk, kc.Provider)
	}

	// Start server
	if err := http.ListenAndServe(":"+port, gw); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

