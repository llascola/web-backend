package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTitleEmpty     = &ErrValidation{Message: "title cannot be empty"}
	ErrContentEmpty   = &ErrValidation{Message: "content cannot be empty"}
	ErrInvalidSlug    = &ErrValidation{Message: "generated slug is invalid"}
	ErrInvalidStatus  = &ErrValidation{Message: "status must be draft or published"}
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

func (s PostStatus) IsValid() bool {
	return s == PostStatusDraft || s == PostStatusPublished
}

type BlogPost struct {
	ID          uuid.UUID
	Title       string
	Slug        string
	Content     string
	Excerpt     string
	Status      PostStatus
	Tags        []string
	AuthorID    uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

var slugRegexp = regexp.MustCompile(`[^a-z0-9]+`)

func generateSlug(title string) string {
	slug := strings.ToLower(strings.TrimSpace(title))
	slug = slugRegexp.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

func NewBlogPost(title, content, excerpt string, status PostStatus, tags []string, authorID uuid.UUID) (*BlogPost, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrTitleEmpty
	}
	if strings.TrimSpace(content) == "" {
		return nil, ErrContentEmpty
	}

	slug := generateSlug(title)
	if slug == "" {
		return nil, ErrInvalidSlug
	}

	if status == "" {
		status = PostStatusDraft
	}
	if !status.IsValid() {
		return nil, ErrInvalidStatus
	}
	if tags == nil {
		tags = []string{}
	}

	now := time.Now()
	post := &BlogPost{
		ID:        uuid.New(),
		Title:     strings.TrimSpace(title),
		Slug:      slug,
		Content:   content,
		Excerpt:   strings.TrimSpace(excerpt),
		Status:    status,
		Tags:      tags,
		AuthorID:  authorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if status == PostStatusPublished {
		post.PublishedAt = &now
	}

	return post, nil
}

func (p *BlogPost) Update(title, content, excerpt *string, status *PostStatus, tags *[]string) error {
	if title != nil {
		trimmed := strings.TrimSpace(*title)
		if trimmed == "" {
			return ErrTitleEmpty
		}
		p.Title = trimmed
		p.Slug = generateSlug(trimmed)
	}
	if content != nil {
		if strings.TrimSpace(*content) == "" {
			return ErrContentEmpty
		}
		p.Content = *content
	}
	if excerpt != nil {
		p.Excerpt = strings.TrimSpace(*excerpt)
	}
	if status != nil {
		if !status.IsValid() {
			return ErrInvalidStatus
		}
		if *status == PostStatusPublished && p.Status != PostStatusPublished {
			now := time.Now()
			p.PublishedAt = &now
		}
		p.Status = *status
	}
	if tags != nil {
		p.Tags = *tags
	}
	p.UpdatedAt = time.Now()
	return nil
}
