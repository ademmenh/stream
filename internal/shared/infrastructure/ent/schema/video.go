package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Video struct {
	ent.Schema
}

func (Video) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("videos"),
	}
}

func (Video) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			MaxLen(255).
			NotEmpty(),
		field.String("description").
			MaxLen(5000).
			Default(""),
		field.Enum("type").
			Values("movie", "series", "documentary"),
		field.Enum("status").
			Values("PendingUpload", "Processing", "Ready", "Failed", "Replacing").
			Default("PendingUpload"),
		field.JSON("qualities", []string{}).
			Default([]string{}),
		field.String("raw_path").
			MaxLen(500).
			Default(""),
		field.String("photo_path").
			MaxLen(500).
			Default(""),
		field.Time("uploaded_at").
			Default(time.Now).
			Immutable(),
	}
}

func (Video) Indexes() []ent.Index {
	return nil
}
