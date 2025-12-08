package inports

import (
	"context"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) error
	RegisterAdmin(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, error)
}
