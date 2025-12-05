package schemamixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

type AtMixin struct {
	mixin.Schema
}

func (AtMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间").
			Annotations(entsql.WithComments(true)),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间").
			Annotations(entsql.WithComments(true)),
	}
}

type AtLogMixin struct {
	mixin.Schema
}

func (AtLogMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间").
			Annotations(entsql.WithComments(true)),
	}
}
