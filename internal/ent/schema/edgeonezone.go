package schema

import (
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

// EdgeOneZone holds the schema definition for the EdgeOneZone entity.
type EdgeOneZone struct {
	ent.Schema
}

// Annotations of the EdgeOneZone.
func (EdgeOneZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edgeone_zone"},
		schema.Comment("EdgeOne 站点"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneZone.
func (EdgeOneZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EdgeOneZone.
func (EdgeOneZone) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("zone_id").
			Comment("站点ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Optional().
			Comment("站点名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("area").
			Values(schematype.GetEdgeOneZoneAreas...).
			Optional().
			Comment("加速区域").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetEdgeOneZoneTypes...).
			Optional().
			Comment("接入类型").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetEdgeOneZoneStatuss...).
			Optional().
			Comment("状态").
			Annotations(entsql.WithComments(true)),

		field.Enum("cname_status").
			Values(schematype.GetEdgeOneZoneCnameStatuss...).
			Optional().
			Comment("CNAME 接入状态").
			Annotations(entsql.WithComments(true)),

		field.Enum("active_status").
			Values(schematype.GetEdgeOneZoneActiveStatuss...).
			Optional().
			Comment("展示状态").
			Annotations(entsql.WithComments(true)),

		field.Bool("paused").
			Optional().
			Comment("是否关闭").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EdgeOneZone.
func (EdgeOneZone) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> EdgeOneZone
		edge.From("config", PlatformConfig.Type).
			Ref("edgeone_zones").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 EdgeOneZone -> EdgeOneZoneAccelerationDomain
		edge.To("accelerationdomains", EdgeOneZoneAccelerationDomain.Type).
			Comment("关联的EdgeOne站点加速域名").
			Annotations(entsql.Annotation{
				// 级联删除：删除 EdgeOneZone 时，其所有 EdgeOneZoneAccelerationDomain 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 一对多 EdgeOneZone -> EdgeOneZoneDNSRecord
		edge.To("dnsrecords", EdgeOneZoneDNSRecord.Type).
			Comment("关联的EdgeOne站点DNS记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 EdgeOneZone 时，其所有 EdgeOneZoneDNSRecord 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 一对多 EdgeOneZone -> EdgeOneZoneOriginGroup
		edge.To("origingroups", EdgeOneZoneOriginGroup.Type).
			Comment("关联的EdgeOne站点源站组").
			Annotations(entsql.Annotation{
				// 级联删除：删除 EdgeOneZone 时，其所有 EdgeOneZoneOriginGroup 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneZone.
func (EdgeOneZone) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("zone_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EdgeOneZone.
func (EdgeOneZone) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EdgeOneZoneLogHook(),
	}
}
