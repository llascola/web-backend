package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/inports"
	"github.com/llascola/web-backend/internal/app/outports"
)

type BlogServiceImpl struct {
	blogRepo outports.BlogPostRepository
}

var _ inports.BlogService = (*BlogServiceImpl)(nil)

func NewBlogService(repo outports.BlogPostRepository) *BlogServiceImpl {
	return &BlogServiceImpl{blogRepo: repo}
}

func (s *BlogServiceImpl) CreatePost(ctx context.Context, title, content, excerpt string, status domain.PostStatus, tags []string, authorID uuid.UUID) (*domain.BlogPost, error) {
	post, err := domain.NewBlogPost(title, content, excerpt, status, tags, authorID)
	if err != nil {
		return nil, err
	}
	if err := s.blogRepo.Save(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *BlogServiceImpl) UpdatePost(ctx context.Context, id uuid.UUID, title, content, excerpt *string, status *domain.PostStatus, tags *[]string) (*domain.BlogPost, error) {
	post, err := s.blogRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := post.Update(title, content, excerpt, status, tags); err != nil {
		return nil, err
	}
	if err := s.blogRepo.Update(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *BlogServiceImpl) DeletePost(ctx context.Context, id uuid.UUID) error {
	return s.blogRepo.Delete(ctx, id)
}

func (s *BlogServiceImpl) GetPublishedBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	post, err := s.blogRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if post.Status != domain.PostStatusPublished {
		return nil, &domain.ErrNotFound{Message: "blog post not found"}
	}
	return post, nil
}

func (s *BlogServiceImpl) ListPublished(ctx context.Context, tag string, page, pageSize int) ([]*domain.BlogPost, int, error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.blogRepo.ListPublished(ctx, tag, page, pageSize)
}

func (s *BlogServiceImpl) ListAll(ctx context.Context, page, pageSize int) ([]*domain.BlogPost, int, error) {
	page, pageSize = normalizePagination(page, pageSize)
	return s.blogRepo.ListAll(ctx, page, pageSize)
}

func (s *BlogServiceImpl) ListTags(ctx context.Context) ([]string, error) {
	return s.blogRepo.ListTags(ctx)
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}
