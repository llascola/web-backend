package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent/user"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
)

type PostgresUserRepository struct {
	client *ent.Client
}

var _ outports.UserRepository = (*PostgresUserRepository)(nil)

func NewUserRepository(client *ent.Client) *PostgresUserRepository {
	return &PostgresUserRepository{client: client}
}

func (r *PostgresUserRepository) Save(ctx context.Context, u *domain.User) error {
	_, err := r.client.User.Create().
		SetID(u.ID).
		SetEmail(u.Email).
		SetPasswordHash(u.PasswordHash).
		SetRole(string(u.Role)).
		SetCreatedAt(u.CreatedAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return &domain.ErrConflict{Message: "user with this email already exists"}
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	q := r.client.User.UpdateOneID(u.ID).
		SetEmail(u.Email).
		SetPasswordHash(u.PasswordHash).
		SetRole(string(u.Role))

	if u.RefreshTokenHash != "" {
		q.SetRefreshTokenHash(u.RefreshTokenHash)
		q.SetRefreshTokenExpiresAt(u.RefreshTokenExpiresAt)
	} else {
		q.ClearRefreshTokenHash()
		q.ClearRefreshTokenExpiresAt()
	}

	err := q.Exec(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return &domain.ErrConflict{Message: "user with this email already exists"}
		}
		return err
	}
	return nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u, err := r.client.User.Query().
		Where(user.Email(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &domain.ErrNotFound{Message: "user not found"}
		}
		return nil, err
	}
	return toDomainUser(u), nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	u, err := r.client.User.Query().
		Where(user.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &domain.ErrNotFound{Message: "user not found"}
		}
		return nil, err
	}
	return toDomainUser(u), nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.User.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return &domain.ErrNotFound{Message: "user not found"}
	}
	return err
}

func toDomainUser(u *ent.User) *domain.User {
	return &domain.User{
		ID:                    u.ID,
		Email:                 u.Email,
		PasswordHash:          u.PasswordHash,
		Role:                  domain.UserRole(u.Role),
		CreatedAt:             u.CreatedAt,
		RefreshTokenHash:      u.RefreshTokenHash,
		RefreshTokenExpiresAt: u.RefreshTokenExpiresAt,
	}
}
