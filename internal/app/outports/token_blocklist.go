package outports

import (
	"context"
	"time"
)

// TokenBlocklist defines the contract for revoking and checking revoked JWT tokens.
type TokenBlocklist interface {
	// Add revokes a token identified by its JTI (JWT ID) until its expiration time.
	Add(ctx context.Context, jti string, expiration time.Time) error

	// IsBlocked returns true if the token identified by JTI has been revoked.
	IsBlocked(ctx context.Context, jti string) (bool, error)
}
