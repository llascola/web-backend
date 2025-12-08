package inports

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
)

type UserService interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}
