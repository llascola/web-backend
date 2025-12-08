package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/outports"
	"github.com/llascola/web-backend/internal/config"
)

type AuthServiceImpl struct {
	userRepo    outports.UserRepository
	jwtKeys     map[string]config.JWTKey
	activeKeyID string
}

var _ inports.AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(repo outports.UserRepository, keys map[string]config.JWTKey, activeKeyID string) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:    repo,
		jwtKeys:     keys,
		activeKeyID: activeKeyID,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, email, password string) error {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return errors.New("user already exists")
	}

	newUser, err := domain.NewUser(email, password, domain.RoleMember)
	if err != nil {
		return err
	}

	return s.userRepo.Save(ctx, newUser)
}

func (s *AuthServiceImpl) RegisterAdmin(ctx context.Context, email, password string) error {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return errors.New("user already exists")
	}

	newUser, err := domain.NewUser(email, password, domain.RoleAdmin)
	if err != nil {
		return err
	}

	return s.userRepo.Save(ctx, newUser)
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || !user.CheckPassword(password) {
		return "", errors.New("invalid credentials")
	}

	keyConfig, ok := s.jwtKeys[s.activeKeyID]
	if !ok {
		return "", errors.New("jwt key not found")
	}

	//TODO : add iat and exp claims
	//TODO : add refresh token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID.String(),
		"role": user.Role,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Minute * 15).Unix(),
	})
	token.Header["kid"] = s.activeKeyID

	tokenString, err := token.SignedString(keyConfig.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
