package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/llascola/web-backend/internal/adapters/driving/rest/openapi"
	"github.com/llascola/web-backend/internal/app/domain"
)

// --- Public ---

func (h *Handler) ListPublishedPosts(ctx *gin.Context, params openapi.ListPublishedPostsParams) {
	tag := ""
	if params.Tag != nil {
		tag = *params.Tag
	}
	page, pageSize := 1, 10
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	posts, total, err := h.blogService.ListPublished(ctx, tag, page, pageSize)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.BlogPostList{
		Posts: toOpenAPIBlogPosts(posts),
		Total: total,
	})
}

func (h *Handler) GetPublishedPost(ctx *gin.Context, slug string) {
	post, err := h.blogService.GetPublishedBySlug(ctx, slug)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toOpenAPIBlogPost(post))
}

func (h *Handler) ListBlogTags(ctx *gin.Context) {
	tags, err := h.blogService.ListTags(ctx)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.TagList{Tags: tags})
}

// --- Admin ---

func (h *Handler) ListAllPosts(ctx *gin.Context, params openapi.ListAllPostsParams) {
	page, pageSize := 1, 10
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	posts, total, err := h.blogService.ListAll(ctx, page, pageSize)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, openapi.BlogPostList{
		Posts: toOpenAPIBlogPosts(posts),
		Total: total,
	})
}

func (h *Handler) CreateBlogPost(ctx *gin.Context) {
	var req openapi.CreateBlogPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		HandleError(ctx, &domain.ErrValidation{Message: "invalid request body"})
		return
	}

	userIDStr, exists := ctx.Get("userID")
	if !exists {
		HandleError(ctx, &domain.ErrUnauthorized{Message: "unauthorized"})
		return
	}
	authorID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		HandleError(ctx, &domain.ErrValidation{Message: "invalid user ID"})
		return
	}

	var tags []string
	if req.Tags != nil {
		tags = *req.Tags
	}

	post, err := h.blogService.CreatePost(
		ctx,
		req.Title,
		req.Content,
		derefString(req.Excerpt),
		domain.PostStatus(req.Status),
		tags,
		authorID,
	)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, toOpenAPIBlogPost(post))
}

func (h *Handler) UpdateBlogPost(ctx *gin.Context, id openapi_types.UUID) {
	var req openapi.UpdateBlogPostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		HandleError(ctx, &domain.ErrValidation{Message: "invalid request body"})
		return
	}

	var status *domain.PostStatus
	if req.Status != nil {
		s := domain.PostStatus(*req.Status)
		status = &s
	}

	post, err := h.blogService.UpdatePost(ctx, id, req.Title, req.Content, req.Excerpt, status, req.Tags)
	if err != nil {
		HandleError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, toOpenAPIBlogPost(post))
}

func (h *Handler) DeleteBlogPost(ctx *gin.Context, id openapi_types.UUID) {
	if err := h.blogService.DeletePost(ctx, id); err != nil {
		HandleError(ctx, err)
		return
	}

	msg := "Blog post deleted"
	ctx.JSON(http.StatusOK, openapi.MessageResponse{Message: &msg})
}

// --- Helpers ---

func toOpenAPIBlogPost(p *domain.BlogPost) openapi.BlogPost {
	result := openapi.BlogPost{
		Id:        p.ID,
		Title:     p.Title,
		Slug:      p.Slug,
		Content:   p.Content,
		Status:    openapi.BlogPostStatus(p.Status),
		Tags:      p.Tags,
		AuthorId:  &p.AuthorID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
	if p.Excerpt != "" {
		result.Excerpt = &p.Excerpt
	}
	if p.PublishedAt != nil {
		result.PublishedAt = p.PublishedAt
	}
	return result
}

func toOpenAPIBlogPosts(posts []*domain.BlogPost) []openapi.BlogPost {
	result := make([]openapi.BlogPost, len(posts))
	for i, p := range posts {
		result[i] = toOpenAPIBlogPost(p)
	}
	return result
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
