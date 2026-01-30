package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EdgeOneZoneOriginGroup holds the schema definition for the EdgeOneZoneOriginGroup entity.
type EdgeOneZoneOriginGroup struct {
	ent.Schema
}

// Annotations of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edgeone_zone_origingroup"},
		schema.Comment("EdgeOne 站点源站组"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点源站组序号").
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
			Values(schematype.GetEdgeOneZoneOriginGroupTypes...).
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
	}
}

// Edges of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EdgeOneZone -> EdgeOneZoneOriginGroup
		edge.From("edgeone_zone", EdgeOneZone.Type).
			Ref("origingroups").
			Field("edgeone_zone_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 EdgeOne 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("edgeone_zone_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EdgeOneZoneOriginGroup.
func (EdgeOneZoneOriginGroup) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EdgeOneZoneOriginGroupLogHook(),
	}
}
