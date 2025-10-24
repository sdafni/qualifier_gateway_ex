package virtualkey

import (
	"fmt"
	"strings"

	"gateway/internal/config"
)

// Service handles virtual key validation and lookup
type Service struct {
	config *config.Config
}

// New creates a new virtual key service
func New(cfg *config.Config) *Service {
	return &Service{config: cfg}
}

// ValidateRequest validates the Authorization header and returns the virtual key
func (s *Service) ValidateRequest(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	// Extract bearer token
	virtualKey := strings.TrimPrefix(authHeader, "Bearer ")
	if virtualKey == authHeader {
		return "", fmt.Errorf("invalid Authorization header format, expected 'Bearer <virtual-key>'")
	}

	// Look up virtual key in config
	if _, exists := s.config.VirtualKeys[virtualKey]; !exists {
		return "", fmt.Errorf("invalid virtual key")
	}

	return virtualKey, nil
}

// GetKeyConfig returns the configuration for a virtual key
func (s *Service) GetKeyConfig(virtualKey string) (config.KeyConfig, bool) {
	keyConfig, exists := s.config.VirtualKeys[virtualKey]
	return keyConfig, exists
}
