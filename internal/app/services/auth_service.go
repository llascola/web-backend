package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/outports"
)

type AuthServiceImpl struct {
	userRepo       outports.UserRepository
	tokenGenerator outports.TokenGenerator
	tokenBlocklist outports.TokenBlocklist
}

var _ inports.AuthService = (*AuthServiceImpl)(nil)

func NewAuthService(repo outports.UserRepository, tokenGenerator outports.TokenGenerator, tokenBlocklist outports.TokenBlocklist) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:       repo,
		tokenGenerator: tokenGenerator,
		tokenBlocklist: tokenBlocklist,
	}
}

func (s *AuthServiceImpl) Register(ctx context.Context, email, password string) error {
	_, err := s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return &domain.ErrConflict{Message: "user already exists"}
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
		return &domain.ErrConflict{Message: "user already exists"}
	}

	newUser, err := domain.NewUser(email, password, domain.RoleAdmin)
	if err != nil {
		return err
	}

	return s.userRepo.Save(ctx, newUser)
}

func (s *AuthServiceImpl) Login(ctx context.Context, email, password string) (string, string, error) {
	genericErr := &domain.ErrUnauthorized{Message: "invalid email or password"}
	user, err := s.userRepo.FindByEmail(ctx, email)

	if err != nil {
		// Dummy check to prevent timing attacks
		domain.CompareDummyPassword(password)
		return "", "", genericErr
	}

	if !user.CheckPassword(password) {
		return "", "", genericErr
	}

	// Generate Refresh Token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	randomStr := base64.RawURLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // 7 days

	if err := user.SetRefreshToken(randomStr, expiresAt); err != nil {
		return "", "", err
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", "", err
	}

	accessToken, err := s.tokenGenerator.GenerateToken(user)
	if err != nil {
		return "", "", err
	}

	refreshTokenPayload := fmt.Sprintf("%s.%s", user.ID.String(), randomStr)
	refreshToken := base64.RawURLEncoding.EncodeToString([]byte(refreshTokenPayload))

	return accessToken, refreshToken, nil
}

func (s *AuthServiceImpl) Logout(ctx context.Context, jti string, exp time.Time) error {
	return s.tokenBlocklist.Add(ctx, jti, exp)
}

func (s *AuthServiceImpl) Refresh(ctx context.Context, refreshTokenParam string) (string, string, error) {
	decodedBytes, err := base64.RawURLEncoding.DecodeString(refreshTokenParam)
	if err != nil {
		return "", "", &domain.ErrValidation{Message: "invalid refresh token format"}
	}
	parts := strings.Split(string(decodedBytes), ".")
	if len(parts) != 2 {
		return "", "", &domain.ErrValidation{Message: "invalid refresh token payload"}
	}

	userID, err := uuid.Parse(parts[0])
	if err != nil {
		return "", "", &domain.ErrValidation{Message: "invalid user id in refresh token"}
	}
	randomStr := parts[1]

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", &domain.ErrUnauthorized{Message: "invalid refresh token"}
	}

	if !user.VerifyRefreshToken(randomStr) {
		return "", "", &domain.ErrUnauthorized{Message: "invalid or expired refresh token"}
	}

	// Token is valid. Rotate refresh token and issue new access token.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	newRandomStr := base64.RawURLEncoding.EncodeToString(b)
	expiresAt := time.Now().Add(time.Hour * 24 * 7)

	if err := user.SetRefreshToken(newRandomStr, expiresAt); err != nil {
		return "", "", err
	}
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", "", err
	}

	accessToken, err := s.tokenGenerator.GenerateToken(user)
	if err != nil {
		return "", "", err
	}

	newRefreshTokenPayload := fmt.Sprintf("%s.%s", user.ID.String(), newRandomStr)
	newRefreshTokenEncoded := base64.RawURLEncoding.EncodeToString([]byte(newRefreshTokenPayload))

	return accessToken, newRefreshTokenEncoded, nil
}
