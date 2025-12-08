package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/outports"
)

type UserServiceImpl struct {
	userRepo outports.UserRepository
}

var _ inports.UserService = (*UserServiceImpl)(nil)

func NewUserService(repo outports.UserRepository) *UserServiceImpl {
	return &UserServiceImpl{
		userRepo: repo,
	}
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.Delete(ctx, userID)
}
