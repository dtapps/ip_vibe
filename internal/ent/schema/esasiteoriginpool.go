package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EsaSiteOriginPool holds the schema definition for the EsaSiteOriginPool entity.
type EsaSiteOriginPool struct {
	ent.Schema
}

// Annotations of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "esa_site_originpool"},
		schema.Comment("ESA 站点源地址池"),
		entsql.WithComments(true),
	}
}

// Mixin of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("ESA 站点源地址池序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("esa_site_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 ESA 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("pool_id").
			Comment("源地址池ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Optional().
			Comment("源地址池名称").
			Annotations(entsql.WithComments(true)),

		field.String("record_name").
			Optional().
			Comment("源地址池分配的域名").
			Annotations(entsql.WithComments(true)),

		field.Bool("enabled").
			Optional().
			Comment("源地址池是否启用").
			Annotations(entsql.WithComments(true)),

		field.JSON("references", json.RawMessage{}).
			Optional().
			Comment("源地址池被引用信息").
			Annotations(entsql.WithComments(true)),

		field.JSON("origins", json.RawMessage{}).
			Optional().
			Comment("源地址池里添加的源站信息").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EsaSite -> EsaSiteOriginPool
		edge.From("esa_site", EsaSite.Type).
			Ref("originpools").
			Field("esa_site_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 ESA 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("esa_site_id").
			Annotations(entsql.WithComments(true)),

		// 按源地址池ID排序
		index.Fields("pool_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EsaSiteOriginPool.
func (EsaSiteOriginPool) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EsaSiteOriginPoolLogHook(),
	}
}
