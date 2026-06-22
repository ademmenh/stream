package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type UserSchema struct {
	ent.Schema
}

func (UserSchema) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Table("users"),
	}
}

func (UserSchema) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			MaxLen(255).
			NotEmpty(),
		field.String("email").
			MaxLen(255).
			Unique().
			NotEmpty(),
		field.String("phone").
			MaxLen(20).
			Optional().
			Nillable(),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty(),
		field.String("role").
			MaxLen(50).
			Default("client"),
		field.Bool("banned").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (UserSchema) Indexes() []ent.Index {
	return nil
}
