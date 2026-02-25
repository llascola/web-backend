package outports

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
)

type BlogPostRepository interface {
	Save(ctx context.Context, post *domain.BlogPost) error
	Update(ctx context.Context, post *domain.BlogPost) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error)
	FindBySlug(ctx context.Context, slug string) (*domain.BlogPost, error)
	ListPublished(ctx context.Context, tag string, page, pageSize int) ([]*domain.BlogPost, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*domain.BlogPost, int, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context) ([]string, error)
}
