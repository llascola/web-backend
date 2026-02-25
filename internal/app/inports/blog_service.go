package inports

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
)

type BlogService interface {
	CreatePost(ctx context.Context, title, content, excerpt string, status domain.PostStatus, tags []string, authorID uuid.UUID) (*domain.BlogPost, error)
	UpdatePost(ctx context.Context, id uuid.UUID, title, content, excerpt *string, status *domain.PostStatus, tags *[]string) (*domain.BlogPost, error)
	DeletePost(ctx context.Context, id uuid.UUID) error
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error)
	ListPublished(ctx context.Context, tag string, page, pageSize int) ([]*domain.BlogPost, int, error)
	ListAll(ctx context.Context, page, pageSize int) ([]*domain.BlogPost, int, error)
	ListTags(ctx context.Context) ([]string, error)
}
