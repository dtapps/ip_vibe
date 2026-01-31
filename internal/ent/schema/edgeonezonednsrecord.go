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

// EdgeOneZoneDNSRecord holds the schema definition for the EdgeOneZoneDNSRecord entity.
type EdgeOneZoneDNSRecord struct {
	ent.Schema
}

// Annotations of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edgeone_zone_dnsrecord"},
		schema.Comment("EdgeOne 站点DNS记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点DNS记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("edgeone_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 EdgeOne 站点序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Comment("DNS 记录 ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Optional().
			Comment("DNS 记录名").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetEdgeOneZoneDNSRecordTypes...).
			Optional().
			Comment("DNS 记录类型").
			Annotations(entsql.WithComments(true)),

		field.String("location").
			Optional().
			Comment("DNS 记录解析线路").
			Annotations(entsql.WithComments(true)),

		field.String("content").
			Optional().
			Comment("DNS 记录内容").
			Annotations(entsql.WithComments(true)),

		field.Int64("ttl").
			Optional().
			Comment("缓存时间").
			Annotations(entsql.WithComments(true)),

		field.Int64("weight").
			Optional().
			Comment("DNS 记录权重").
			Annotations(entsql.WithComments(true)),

		field.Int64("priority").
			Optional().
			Comment("MX 记录优先级").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetEdgeOneZoneDNSRecordStatuss...).
			Optional().
			Comment("DNS 记录解析状态").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EdgeOneZone -> EdgeOneZoneDNSRecord
		edge.From("edgeone_zone", EdgeOneZone.Type).
			Ref("dnsrecords").
			Field("edgeone_zone_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 EdgeOne 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("edgeone_zone_id").
			Annotations(entsql.WithComments(true)),

		// 按DNS记录ID排序
		index.Fields("record_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EdgeOneZoneDNSRecord.
func (EdgeOneZoneDNSRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EdgeOneZoneDNSRecordLogHook(),
	}
}
