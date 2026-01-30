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
			Immutable().
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

type SyncAtMixin struct {
	mixin.Schema
}

func (SyncAtMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("remote_created_at").
			Optional().
			Immutable().
			Comment("远程创建时间").
			Annotations(entsql.WithComments(true)),

		field.Time("sync_at").
			Optional().
			Comment("最近一次同步时间").
			Annotations(entsql.WithComments(true)),
	}
}
