package services_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports/mocks"
	"github.com/llascola/web-backend/internal/app/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserService_GetProfile(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Returns User Successfully", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewUserService(mockRepo)

		expectedUser := &domain.User{ID: userID, Email: "test@example.com"}
		mockRepo.On("FindByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetProfile(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns Error on DB Failure", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewUserService(mockRepo)

		mockRepo.On("FindByID", ctx, userID).Return(nil, &domain.ErrInternal{Message: "db connection failure"})

		user, err := service.GetProfile(ctx, userID)

		var target *domain.ErrInternal
		require.ErrorAs(t, err, &target)
		assert.Nil(t, user)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("Deletes User Successfully", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewUserService(mockRepo)

		mockRepo.On("Delete", ctx, userID).Return(nil)

		err := service.DeleteUser(ctx, userID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Returns Error on Deletion Failure", func(t *testing.T) {
		mockRepo := mocks.NewMockUserRepository(t)
		service := services.NewUserService(mockRepo)

		mockRepo.On("Delete", ctx, userID).Return(&domain.ErrNotFound{Message: "user not found"})

		err := service.DeleteUser(ctx, userID)

		var target *domain.ErrNotFound
		require.ErrorAs(t, err, &target)
		mockRepo.AssertExpectations(t)
	})
}
