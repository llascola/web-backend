package domain_test

import (
	"testing"
	"time"

	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		password    string
		role        domain.UserRole
		expectedErr error
	}{
		{
			name:        "Valid Member User",
			email:       "test@example.com",
			password:    "securePassword123!",
			role:        domain.RoleMember,
			expectedErr: nil,
		},
		{
			name:        "Valid Admin User",
			email:       "admin@example.com",
			password:    "securePassword123!",
			role:        domain.RoleAdmin,
			expectedErr: nil,
		},
		{
			name:        "Default Role Fallback",
			email:       "default@example.com",
			password:    "securePassword123!",
			role:        "", // Should default to RoleMember
			expectedErr: nil,
		},
		{
			name:        "Password Too Short",
			email:       "short@example.com",
			password:    "short",
			role:        domain.RoleMember,
			expectedErr: domain.ErrPasswordWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := domain.NewUser(tt.email, tt.password, tt.role)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				require.NotNil(t, user)

				assert.Equal(t, tt.email, user.Email)
				assert.NotEmpty(t, user.ID)
				assert.NotEmpty(t, user.PasswordHash)
				assert.NotEqual(t, tt.password, user.PasswordHash, "Password should be hashed")
				assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)

				// Verify Default Role behavior
				if tt.role == "" {
					assert.Equal(t, domain.RoleMember, user.Role)
				} else {
					assert.Equal(t, tt.role, user.Role)
				}
			}
		})
	}
}

func TestVerifyRefreshToken(t *testing.T) {
	user, err := domain.NewUser("test@example.com", "secure12345", domain.RoleMember)
	require.NoError(t, err)

	validToken := "my-secure-long-refresh-token"
	validExpiration := time.Now().Add(time.Hour)

	err = user.SetRefreshToken(validToken, validExpiration)
	require.NoError(t, err)
	require.NotEmpty(t, user.RefreshTokenHash)

	tests := []struct {
		name     string
		token    string
		mutateFn func(*domain.User)
		valid    bool
	}{
		{
			name:  "Valid Token Match",
			token: validToken,
			valid: true,
		},
		{
			name:  "Invalid Token Mismatch",
			token: "wrong-token",
			valid: false,
		},
		{
			name:  "Expired Token",
			token: validToken,
			mutateFn: func(u *domain.User) {
				u.RefreshTokenExpiresAt = time.Now().Add(-time.Hour)
			},
			valid: false,
		},
		{
			name:  "Missing Hash",
			token: validToken,
			mutateFn: func(u *domain.User) {
				u.RefreshTokenHash = ""
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clone user to not mutate across tests implicitly
			testUser := *user
			if tt.mutateFn != nil {
				tt.mutateFn(&testUser)
			}

			isValid := testUser.VerifyRefreshToken(tt.token)
			assert.Equal(t, tt.valid, isValid)
		})
	}
}
