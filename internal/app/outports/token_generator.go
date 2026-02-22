package outports

import "github.com/llascola/web-backend/internal/app/domain"

// TokenGenerator defines the contract for creating authentication tokens (e.g., JWT).
type TokenGenerator interface {
	GenerateToken(user *domain.User) (string, error)
}
