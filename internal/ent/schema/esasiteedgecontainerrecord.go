package schema

import (
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EsaSiteEdgeContainerRecord holds the schema definition for the EsaSiteEdgeContainerRecord entity.
type EsaSiteEdgeContainerRecord struct {
	ent.Schema
}

// Annotations of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "esa_site_edge_container_record"},
		schema.Comment("ESA 站点边缘容器记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("ESA 站点边缘容器记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("esa_site_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 ESA 站点序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_name").
			Immutable().
			Comment("记录名称").
			Annotations(entsql.WithComments(true)),

		field.String("record_cname").
			Optional().
			Comment("记录的 CNAME").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EsaSite -> EsaSiteEdgeContainerRecord
		edge.From("esa_site", EsaSite.Type).
			Ref("edge_container_records").
			Field("esa_site_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 ESA 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("esa_site_id").
			Annotations(entsql.WithComments(true)),

		// 按记录名称排序
		index.Fields("record_name").
			Annotations(entsql.WithComments(true)),

		// 按站点ID+记录名称排序
		index.Fields("esa_site_id", "record_name").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EsaSiteEdgeContainerRecord.
func (EsaSiteEdgeContainerRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EsaSiteEdgeContainerRecordLogHook(),
	}
}
