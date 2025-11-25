package schema

import (
	"vibe_ip/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
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
			StorageKey("id").
			Immutable().
			Comment("报告序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Int64("device_id").
			Immutable().
			Comment("设备序号").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v4").
			Optional().
			Immutable().
			Comment("IP地址V4").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v6").
			Optional().
			Immutable().
			Comment("IP地址V6").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DeviceIpReport.
func (DeviceIpReport) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 Device -> DeviceIpReport
		edge.From("device", Device.Type).
			Ref("ip_reports").
			Field("device_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的设备").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DeviceIpReport.
func (DeviceIpReport) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("device_id").
			Annotations(entsql.WithComments(true)),
	}
}
