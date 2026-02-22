package services_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports/mocks"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Refresh(t *testing.T) {
	ctx := context.Background()

	// Helper to generate a valid domain User and its corresponding opaque refresh token payload
	setupValidUser := func() (*domain.User, string) {
		user, _ := domain.NewUser("test@example.com", "secure123", domain.RoleMember)

		// Create a realistic stored refresh token configuration
		randomStr := "valid-random-string"
		expiresAt := time.Now().Add(time.Hour * 24)
		_ = user.SetRefreshToken(randomStr, expiresAt)

		// Create the token string the client would send: base64(userID.randomStr)
		payload := fmt.Sprintf("%s.%s", user.ID.String(), randomStr)
		clientToken := base64.RawURLEncoding.EncodeToString([]byte(payload))

		return user, clientToken
	}

	t.Run("Success Rotating Tokens", func(t *testing.T) {
		user, clientToken := setupValidUser()

		mockRepo := mocks.NewMockUserRepository(t)
		mockTokenGen := mocks.NewMockTokenGenerator(t)
		mockBlocklist := mocks.NewMockTokenBlocklist(t)

		service := services.NewAuthService(mockRepo, mockTokenGen, mockBlocklist)

		// Expectations
		mockRepo.On("FindByID", ctx, user.ID).Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
			updatedUser := args.Get(1).(*domain.User)
			assert.NotEqual(t, "valid-random-string", updatedUser.RefreshTokenHash, "Refresh token should have been rotated")
		})
		mockTokenGen.On("GenerateToken", user).Return("new_access_token_jwt", nil)

		// Execution
		accessToken, refreshToken, err := service.Refresh(ctx, clientToken)

		// Assertions
		require.NoError(t, err)
		assert.Equal(t, "new_access_token_jwt", accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.NotEqual(t, clientToken, refreshToken, "Refresh token must change on rotation")

		mockRepo.AssertExpectations(t)
		mockTokenGen.AssertExpectations(t)
	})

	t.Run("Fails with Invalid Base64 Format", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil) // Dependencies unused

		_, _, err := service.Refresh(ctx, "invalid-base64!!!")
		var target *domain.ErrValidation
		require.ErrorAs(t, err, &target)
	})

	t.Run("Fails with Missing User ID payload", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		badPayload := base64.RawURLEncoding.EncodeToString([]byte("just-a-string-without-dot"))
		_, _, err := service.Refresh(ctx, badPayload)
		var target *domain.ErrValidation
		require.ErrorAs(t, err, &target)
	})

	t.Run("Fails if User Not Found in DB", func(t *testing.T) {
		user, clientToken := setupValidUser()

		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByID", ctx, user.ID).Return(nil, errors.New("db error"))

		_, _, err := service.Refresh(ctx, clientToken)
		var target *domain.ErrUnauthorized
		require.ErrorAs(t, err, &target)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fails on Expired or Reused Token (Hash Mismatch)", func(t *testing.T) {
		user, _ := setupValidUser()

		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByID", ctx, user.ID).Return(user, nil)

		// Send a structurally valid token acting as a reused/stolen token
		stolenPayload := fmt.Sprintf("%s.wrong-random-string", user.ID.String())
		stolenToken := base64.RawURLEncoding.EncodeToString([]byte(stolenPayload))

		_, _, err := service.Refresh(ctx, stolenToken)
		var target *domain.ErrUnauthorized
		require.ErrorAs(t, err, &target)

		// Assert Repo.Update and TokenGen.GenerateToken were never called
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("Success Register Member", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByEmail", ctx, "new@example.com").Return(nil, errors.New("not found"))
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(1).(*domain.User)
			assert.Equal(t, domain.RoleMember, user.Role)
			assert.Equal(t, "new@example.com", user.Email)
		})

		err := service.Register(ctx, "new@example.com", "secure123!")
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success Register Admin", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByEmail", ctx, "admin@example.com").Return(nil, errors.New("not found"))
		mockRepo.On("Save", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(1).(*domain.User)
			assert.Equal(t, domain.RoleAdmin, user.Role)
		})

		err := service.RegisterAdmin(ctx, "admin@example.com", "secure123!")
		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fails if Email Already Exists", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByEmail", ctx, "existing@example.com").Return(&domain.User{}, nil)
		// Save should NOT be called

		err := service.Register(ctx, "existing@example.com", "secure123!")
		var target *domain.ErrConflict
		require.ErrorAs(t, err, &target)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("Success Login and Token Generation", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		mockTokenGen := mocks.NewMockTokenGenerator(t)
		service := services.NewAuthService(mockRepo, mockTokenGen, nil)

		user, _ := domain.NewUser("test@example.com", "correctPassword", domain.RoleMember)

		mockRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
			updatedUser := args.Get(1).(*domain.User)
			assert.NotEmpty(t, updatedUser.RefreshTokenHash)
			assert.False(t, updatedUser.RefreshTokenExpiresAt.IsZero())
		})
		mockTokenGen.On("GenerateToken", user).Return("access_jwt_abc123", nil)

		accessToken, refreshToken, err := service.Login(ctx, "test@example.com", "correctPassword")

		require.NoError(t, err)
		assert.Equal(t, "access_jwt_abc123", accessToken)
		assert.NotEmpty(t, refreshToken)

		// Ensure the returned refresh token contains the UUID part
		decodedRT, err := base64.RawURLEncoding.DecodeString(refreshToken)
		require.NoError(t, err)
		parts := strings.Split(string(decodedRT), ".")
		assert.Len(t, parts, 2)
		assert.Equal(t, user.ID.String(), parts[0])

		mockRepo.AssertExpectations(t)
		mockTokenGen.AssertExpectations(t)
	})

	t.Run("Fails on Invalid Email", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		mockRepo.On("FindByEmail", ctx, "bad@example.com").Return(nil, errors.New("db error"))

		accessToken, refreshToken, err := service.Login(ctx, "bad@example.com", "anyPassword")

		var target *domain.ErrUnauthorized
		require.ErrorAs(t, err, &target)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Fails on Wrong Password", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewAuthService(mockRepo, nil, nil)

		user, _ := domain.NewUser("test@example.com", "correctPassword", domain.RoleMember)
		mockRepo.On("FindByEmail", ctx, "test@example.com").Return(user, nil)

		accessToken, refreshToken, err := service.Login(ctx, "test@example.com", "wrongPassword")

		var target *domain.ErrUnauthorized
		require.ErrorAs(t, err, &target)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("Passes through blocklist", func(t *testing.T) {
		mockBlocklist := mocks.NewMockTokenBlocklist(t)
		service := services.NewAuthService(nil, nil, mockBlocklist)

		exp := time.Now().Add(15 * time.Minute)
		mockBlocklist.On("Add", ctx, "jti-123456", exp).Return(nil)

		err := service.Logout(ctx, "jti-123456", exp)
		require.NoError(t, err)

		mockBlocklist.AssertExpectations(t)
	})
}
