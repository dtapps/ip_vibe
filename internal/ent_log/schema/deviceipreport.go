package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schemavalidate"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DeviceIpReport holds the schema definition for the DeviceIpReport entity.
type DeviceIpReport struct {
	ent.Schema
}

// Annotations of the DeviceIpReport.
func (DeviceIpReport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_ip_report"},
		schema.Comment("设备IP报告"),
		entsql.WithComments(true),
	}
}

// Mixin of the DeviceIpReport.
func (DeviceIpReport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the DeviceIpReport.
func (DeviceIpReport) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("IP报告序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("device_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 Device 序号").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v4").
			Validate(schemavalidate.ValidateIPV4).
			Optional().
			Immutable().
			Comment("IP地址V4").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v6").
			Validate(schemavalidate.ValidateIPV6).
			Optional().
			Immutable().
			Comment("IP地址V6").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DeviceIpReport.
func (DeviceIpReport) Edges() []ent.Edge {
	return nil
}

// Indexes of the DeviceIpReport.
func (DeviceIpReport) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间降序排序，便于查询最新记录
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按关联 Device 序号排序，便于过滤记录
		index.Fields("device_id").
			Annotations(entsql.WithComments(true)),

		// 复合索引：关联 Device 序号 + 创建时间，用于高效时间范围查询
		index.Fields("device_id", "created_at").
			Annotations(entsql.WithComments(true)),
	}
}
