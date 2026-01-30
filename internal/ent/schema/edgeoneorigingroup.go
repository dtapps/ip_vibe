package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EdgeOneOriginGroup holds the schema definition for the EdgeOneOriginGroup entity.
type EdgeOneOriginGroup struct {
	ent.Schema
}

// Annotations of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edgeone_origin_group"},
		schema.Comment("EdgeOne 源站组"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 源站组序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("edgeone_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 EdgeOne 站点序号").
			Annotations(entsql.WithComments(true)),

		field.String("group_id").
			Optional().
			Comment("源站组ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("源站组名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetEdgeOneOriginGroupTypes...).
			Optional().
			Comment("源站组类型").
			Annotations(entsql.WithComments(true)),

		field.JSON("records", json.RawMessage{}).
			Optional().
			Comment("源站记录信息").
			Annotations(entsql.WithComments(true)),

		field.JSON("references", json.RawMessage{}).
			Optional().
			Comment("源站组被引用实例列表").
			Annotations(entsql.WithComments(true)),

		field.Time("remote_created_at").
			Immutable().
			Comment("源站组创建时间").
			Annotations(entsql.WithComments(true)),

		field.Time("sync_at").
			Optional().
			Comment("最近一次同步时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EdgeOneZone -> EdgeOneOriginGroup
		edge.From("edgeone_zone", EdgeOneZone.Type).
			Ref("origin_groups").
			Field("edgeone_zone_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 EdgeOne 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("edgeone_zone_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EdgeOneOriginGroup.
func (EdgeOneOriginGroup) Hooks() []ent.Hook {
	return []ent.Hook{
		// schemahook.EdgeOneOriginGroupLogHook(),
	}
}
