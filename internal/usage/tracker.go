package usage

import (
	"fmt"
	"sync"
	"time"
)

// VirtualKeyUsage tracks usage for a single virtual key
type VirtualKeyUsage struct {
	requestCount int
	windowStart  time.Time
}

// Tracker manages per-virtual-key usage quotas
type Tracker struct {
	mu                 sync.RWMutex
	usage              map[string]*VirtualKeyUsage
	maxRequestsPerHour int
}

// New creates a new usage tracker with per-key hourly quotas
func New(maxRequestsPerHour int) *Tracker {
	return &Tracker{
		usage:              make(map[string]*VirtualKeyUsage),
		maxRequestsPerHour: maxRequestsPerHour,
	}
}

// getOrCreateUsage gets or initializes usage data for a virtual key
// Must be called with lock held
func (t *Tracker) getOrCreateUsage(virtualKey string) *VirtualKeyUsage {
	if usage, exists := t.usage[virtualKey]; exists {
		return usage
	}

	// Initialize new virtual key usage
	t.usage[virtualKey] = &VirtualKeyUsage{
		requestCount: 0,
		windowStart:  time.Now(),
	}
	return t.usage[virtualKey]
}

// CheckQuota checks if the request is within quota for the given virtual key
// Returns an error if quota is exceeded
func (t *Tracker) CheckQuota(virtualKey string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage := t.getOrCreateUsage(virtualKey)

	// Reset counter if we've moved to a new hour window
	if time.Since(usage.windowStart) >= time.Hour {
		usage.requestCount = 0
		usage.windowStart = time.Now()
	}

	// Check if quota exceeded
	if usage.requestCount >= t.maxRequestsPerHour {
		return fmt.Errorf("quota exceeded: %d requests per hour limit reached", t.maxRequestsPerHour)
	}

	return nil
}

// RecordRequest increments the request counter for the given virtual key
func (t *Tracker) RecordRequest(virtualKey string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	usage := t.getOrCreateUsage(virtualKey)

	// Reset counter if we've moved to a new hour window
	if time.Since(usage.windowStart) >= time.Hour {
		usage.requestCount = 0
		usage.windowStart = time.Now()
	}

	usage.requestCount++
}

// GetStats returns current usage statistics for a specific virtual key
func (t *Tracker) GetStats(virtualKey string) (requestCount int, maxRequests int, windowStart time.Time) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if usage, exists := t.usage[virtualKey]; exists {
		return usage.requestCount, t.maxRequestsPerHour, usage.windowStart
	}

	// Return zero values for unknown key
	return 0, t.maxRequestsPerHour, time.Time{}
}
