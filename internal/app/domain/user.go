package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail = &ErrValidation{Message: "invalid email format"}
	ErrPasswordWeak = &ErrValidation{Message: "password must be at least 8 characters"}
)

type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type User struct {
	ID                    uuid.UUID
	Email                 string
	PasswordHash          string
	Role                  UserRole
	CreatedAt             time.Time
	RefreshTokenHash      string
	RefreshTokenExpiresAt time.Time
}

func NewUser(email, password string, role UserRole) (*User, error) {
	if len(password) < 8 {
		return nil, ErrPasswordWeak
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	if role == "" {
		role = RoleMember
	}

	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         role,
		CreatedAt:    time.Now(),
	}, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) SetRefreshToken(token string, expiresAt time.Time) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), 10) // Lower cost fine for high entropy tokens
	if err != nil {
		return err
	}
	u.RefreshTokenHash = string(hash)
	u.RefreshTokenExpiresAt = expiresAt
	return nil
}

func (u *User) VerifyRefreshToken(token string) bool {
	if u.RefreshTokenHash == "" || time.Now().After(u.RefreshTokenExpiresAt) {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.RefreshTokenHash), []byte(token)) == nil
}

// CompareDummyPassword runs a bcrypt comparison against a dummy hash to normalize response times
// for login attempts with valid vs invalid emails.
func CompareDummyPassword(password string) {
	// A valid pre-computed bcrypt hash (cost 12) for an empty string or generic value
	dummyHash := []byte("$2a$12$DUMMYDUMMYDUMMYDUMMYDUMMYDUMMYDUMMYDUMMYDUMMYDUMMYDUM")
	_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
}
