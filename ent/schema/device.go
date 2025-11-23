package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Device holds the schema definition for the Device entity.
type Device struct {
	ent.Schema
}

// Annotations of the Device.
func (Device) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device"},
		schema.Comment("设备"),
		entsql.WithComments(true),
	}
}

// Fields of the Device.
func (Device) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("设备序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.String("code").
			Unique().
			Immutable().
			Comment("设备编号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("设备名称").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v4").
			Optional().
			Comment("IP地址V4").
			Annotations(entsql.WithComments(true)),

		field.String("ip_v6").
			Optional().
			Comment("IP地址V6").
			Annotations(entsql.WithComments(true)),

		field.Time("last_reported_at").
			Optional(). // 允许为空（新设备尚未上报）
			Comment("最后上报时间").
			Annotations(entsql.WithComments(true)),

		field.Time("last_heartbeat_at").
			Optional(). // 允许为空（新设备尚未上报）
			Comment("最后心跳时间").
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间").
			Annotations(entsql.WithComments(true)),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 Device -> DeviceIpReport
		edge.To("ip_reports", DeviceIpReport.Type).
			Comment("IP报告").
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 多对多：Device <-> IPGroup
		edge.To("ip_groups", IpGroup.Type).
			Through("ip_group_device", IpGroupDevice.Type).
			Comment("IP组").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the Device.
func (Device) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}
