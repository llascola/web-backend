package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/llascola/web-backend/internal/app/outports"
)

// InMemoryTokenBlocklist is a thread-safe in-memory adapter for TokenBlocklist outport.
// It is intended for E2E testing and local development, not for production use.
type InMemoryTokenBlocklist struct {
	mu     sync.RWMutex
	tokens map[string]time.Time
}

var _ outports.TokenBlocklist = (*InMemoryTokenBlocklist)(nil)

// NewInMemoryTokenBlocklist creates a new in-memory blocklist.
func NewInMemoryTokenBlocklist() *InMemoryTokenBlocklist {
	return &InMemoryTokenBlocklist{
		tokens: make(map[string]time.Time),
	}
}

// Add inserts a token (JTI) into the blocklist with an expiration time.
func (b *InMemoryTokenBlocklist) Add(ctx context.Context, jti string, exp time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.tokens[jti] = exp
	return nil
}

// IsBlocked checks if a token exists in the blocklist and if it's still unexpired.
func (b *InMemoryTokenBlocklist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	exp, exists := b.tokens[jti]
	if !exists {
		return false, nil
	}

	// If the current time is past the expiration time, it "evicts" logically
	// Note: We don't clean it up here to keep the read strictly RLock,
	// but it counts as not blocked because the original token expired anyway.
	if time.Now().After(exp) {
		return false, nil
	}

	return true, nil
}

// Cleanup is a helper method (not in the interface) that tests can call
// to wipe the state between test cases.
func (b *InMemoryTokenBlocklist) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens = make(map[string]time.Time)
}

// SimulateError is an optional hook tests can use to simulate a Redis failure
// Not implemented fully here, but can be added if tests require simulating disconnects.
var ErrSimulatedDatabaseMissing = errors.New("simulated database missing")
