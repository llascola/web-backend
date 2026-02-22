package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/llascola/web-backend/internal/config"
)

// JWTGenerator implements the TokenGenerator port using jwt/v5.
type JWTGenerator struct {
	keys        map[string]config.JWTKey
	activeKeyID string
}

var _ outports.TokenGenerator = (*JWTGenerator)(nil)

// NewJWTGenerator creates a new JWT generator driven adapter.
func NewJWTGenerator(keys map[string]config.JWTKey, activeKeyID string) *JWTGenerator {
	return &JWTGenerator{
		keys:        keys,
		activeKeyID: activeKeyID,
	}
}

// GenerateToken creates a signed JWT for the given user.
func (g *JWTGenerator) GenerateToken(user *domain.User) (string, error) {
	keyConfig, ok := g.keys[g.activeKeyID]
	if !ok {
		return "", errors.New("jwt key not found")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
		"iss":  "portfolio-api",
		"aud":  "portfolio-frontend",
		"jti":  uuid.New().String(),
	})
	token.Header["kid"] = g.activeKeyID

	tokenString, err := token.SignedString(keyConfig.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
