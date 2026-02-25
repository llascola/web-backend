package postgres

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent/blogpost"
	"github.com/llascola/web-backend/internal/adapters/driven/repository/ent/tag"
	"github.com/llascola/web-backend/internal/app/domain"
	"github.com/llascola/web-backend/internal/app/outports"
)

type PostgresBlogPostRepository struct {
	client *ent.Client
}

var _ outports.BlogPostRepository = (*PostgresBlogPostRepository)(nil)

func NewBlogPostRepository(client *ent.Client) *PostgresBlogPostRepository {
	return &PostgresBlogPostRepository{client: client}
}

func (r *PostgresBlogPostRepository) Save(ctx context.Context, p *domain.BlogPost) error {
	tagIDs, err := r.ensureTags(ctx, p.Tags)
	if err != nil {
		return err
	}

	builder := r.client.BlogPost.Create().
		SetID(p.ID).
		SetTitle(p.Title).
		SetSlug(p.Slug).
		SetContent(p.Content).
		SetExcerpt(p.Excerpt).
		SetStatus(blogpost.Status(p.Status)).
		SetAuthorID(p.AuthorID).
		SetCreatedAt(p.CreatedAt).
		SetUpdatedAt(p.UpdatedAt).
		AddTagIDs(tagIDs...)

	if p.PublishedAt != nil {
		builder.SetPublishedAt(*p.PublishedAt)
	}

	_, err = builder.Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return &domain.ErrConflict{Message: "a blog post with this slug already exists"}
		}
		return err
	}
	return nil
}

func (r *PostgresBlogPostRepository) Update(ctx context.Context, p *domain.BlogPost) error {
	tagIDs, err := r.ensureTags(ctx, p.Tags)
	if err != nil {
		return err
	}

	builder := r.client.BlogPost.UpdateOneID(p.ID).
		SetTitle(p.Title).
		SetSlug(p.Slug).
		SetContent(p.Content).
		SetExcerpt(p.Excerpt).
		SetStatus(blogpost.Status(p.Status)).
		SetUpdatedAt(p.UpdatedAt).
		ClearTags().
		AddTagIDs(tagIDs...)

	if p.PublishedAt != nil {
		builder.SetPublishedAt(*p.PublishedAt)
	} else {
		builder.ClearPublishedAt()
	}

	err = builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &domain.ErrNotFound{Message: "blog post not found"}
		}
		if ent.IsConstraintError(err) {
			return &domain.ErrConflict{Message: "a blog post with this slug already exists"}
		}
		return err
	}
	return nil
}

func (r *PostgresBlogPostRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.BlogPost, error) {
	p, err := r.client.BlogPost.Query().
		Where(blogpost.IDEQ(id)).
		WithTags().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &domain.ErrNotFound{Message: "blog post not found"}
		}
		return nil, err
	}
	return toDomainBlogPost(p), nil
}

func (r *PostgresBlogPostRepository) FindBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p, err := r.client.BlogPost.Query().
		Where(blogpost.SlugEQ(slug)).
		WithTags().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, &domain.ErrNotFound{Message: "blog post not found"}
		}
		return nil, err
	}
	return toDomainBlogPost(p), nil
}

func (r *PostgresBlogPostRepository) ListPublished(ctx context.Context, tagFilter string, page, pageSize int) ([]*domain.BlogPost, int, error) {
	query := r.client.BlogPost.Query().
		Where(blogpost.StatusEQ(blogpost.StatusPublished))

	if tagFilter != "" {
		query = query.Where(blogpost.HasTagsWith(tag.NameEQ(tagFilter)))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	posts, err := query.
		WithTags().
		Order(blogpost.ByPublishedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return toDomainBlogPosts(posts), total, nil
}

func (r *PostgresBlogPostRepository) ListAll(ctx context.Context, page, pageSize int) ([]*domain.BlogPost, int, error) {
	total, err := r.client.BlogPost.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	posts, err := r.client.BlogPost.Query().
		WithTags().
		Order(blogpost.ByUpdatedAt(sql.OrderDesc())).
		Offset(offset).
		Limit(pageSize).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return toDomainBlogPosts(posts), total, nil
}

func (r *PostgresBlogPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.client.BlogPost.DeleteOneID(id).Exec(ctx)
	if ent.IsNotFound(err) {
		return &domain.ErrNotFound{Message: "blog post not found"}
	}
	return err
}

func (r *PostgresBlogPostRepository) ListTags(ctx context.Context) ([]string, error) {
	tags, err := r.client.Tag.Query().
		Where(tag.HasPostsWith(blogpost.StatusEQ(blogpost.StatusPublished))).
		Order(tag.ByName()).
		All(ctx)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names, nil
}

// ensureTags finds existing tags by name and creates any that are missing,
// returning all tag IDs ready to attach to a BlogPost edge.
func (r *PostgresBlogPostRepository) ensureTags(ctx context.Context, names []string) ([]uuid.UUID, error) {
	if len(names) == 0 {
		return nil, nil
	}

	existing, err := r.client.Tag.Query().
		Where(tag.NameIn(names...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]uuid.UUID, len(existing))
	for _, t := range existing {
		existingMap[t.Name] = t.ID
	}

	ids := make([]uuid.UUID, 0, len(names))
	for _, name := range names {
		if id, ok := existingMap[name]; ok {
			ids = append(ids, id)
			continue
		}
		t, err := r.client.Tag.Create().SetName(name).Save(ctx)
		if err != nil {
			if ent.IsConstraintError(err) {
				// Another request created it concurrently; fetch it.
				t, err = r.client.Tag.Query().Where(tag.NameEQ(name)).Only(ctx)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		ids = append(ids, t.ID)
	}
	return ids, nil
}

func toDomainBlogPost(p *ent.BlogPost) *domain.BlogPost {
	tags := make([]string, len(p.Edges.Tags))
	for i, t := range p.Edges.Tags {
		tags[i] = t.Name
	}
	return &domain.BlogPost{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		Excerpt:     p.Excerpt,
		Status:      domain.PostStatus(p.Status),
		Tags:        tags,
		AuthorID:    p.AuthorID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		PublishedAt: p.PublishedAt,
	}
}

func toDomainBlogPosts(posts []*ent.BlogPost) []*domain.BlogPost {
	result := make([]*domain.BlogPost, len(posts))
	for i, p := range posts {
		result[i] = toDomainBlogPost(p)
	}
	return result
}
