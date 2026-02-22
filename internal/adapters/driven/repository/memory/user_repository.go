package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
)

type InMemoryUserRepository struct {
	users map[uuid.UUID]*domain.User
	mu    sync.RWMutex
}

var _ outports.UserRepository = (*InMemoryUserRepository)(nil)

func NewUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *InMemoryUserRepository) Save(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) Update(ctx context.Context, user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.users[user.ID]; !exists {
		return &domain.ErrNotFound{Message: "user not found for update"}
	}
	r.users[user.ID] = user
	return nil
}

func (r *InMemoryUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, &domain.ErrNotFound{Message: "user not found"}
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, exists := r.users[id]
	if !exists {
		return nil, &domain.ErrNotFound{Message: "user not found"}
	}
	return user, nil
}

func (r *InMemoryUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}
