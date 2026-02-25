package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// BlogPost holds the schema definition for the BlogPost entity.
type BlogPost struct {
	ent.Schema
}

// Fields of the BlogPost.
func (BlogPost) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New),
		field.String("title").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.Text("content").
			NotEmpty(),
		field.String("excerpt").
			Optional().
			Default(""),
		field.Enum("status").
			Values("draft", "published"),
		field.UUID("author_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("published_at").
			Optional().
			Nillable(),
	}
}

// Edges of the BlogPost.
func (BlogPost) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("tags", Tag.Type),
	}
}
