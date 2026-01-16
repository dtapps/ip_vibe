package schema

import (
	"vibe_ip/internal/ent/schema/schemahook"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Cache holds the schema definition for the Cache entity.
type Cache struct {
	ent.Schema
}

// Annotations of the Cache.
func (Cache) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cache"},
		schema.Comment("缓存数据"),
		entsql.WithComments(true),
	}
}

// Mixin of the Cache.
func (Cache) Mixin() []ent.Mixin {
	return []ent.Mixin{}
}

// Fields of the Cache.
func (Cache) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("缓存记录序号").
			Annotations(entsql.WithComments(true)),

		field.String("key").
			Unique().
			Immutable().
			Comment("缓存键").
			Annotations(entsql.WithComments(true)),

		field.Bytes("value").
			Optional().
			Comment("缓存值").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the Cache.
func (Cache) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the Cache.
func (Cache) Indexes() []ent.Index {
	return []ent.Index{
		// 按缓存键排序
		index.Fields("key").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the Cache.
func (Cache) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.CacheLogHook(),
	}
}
