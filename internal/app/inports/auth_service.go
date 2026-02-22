package inports

import (
	"context"
	"time"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) error
	RegisterAdmin(ctx context.Context, email, password string) error
	Login(ctx context.Context, email, password string) (string, string, error)
	Refresh(ctx context.Context, refreshTokenParam string) (string, string, error)
	Logout(ctx context.Context, jti string, exp time.Time) error
}
