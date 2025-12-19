package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

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

// Mixin of the Device.
func (Device) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
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

		field.String("description").
			Optional().
			Comment("设备描述").
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
	}
}

// Edges of the Device.
func (Device) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 Device -> DeviceIpReport
		edge.To("ip_reports", DeviceIpReport.Type).
			Comment("关联的设备IP报告").
			Annotations(entsql.Annotation{
				// 级联删除：删除 Device 时，其所有 DeviceIpReport 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 多对多：Device <-> IPGroup
		edge.To("ip_groups", IpGroup.Type).
			Through("ip_group_device", IpGroupDevice.Type).
			Comment("关联的IP组").
			Annotations(entsql.WithComments(true)),

		edge.To("port_forwards", DevicePortForward.Type).
			Comment("设备的端口转发规则").
			Annotations(entsql.Annotation{
				// 级联删除：删除 Device 时，其所有 DevicePortForward 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the Device.
func (Device) Indexes() []ent.Index {
	return []ent.Index{
		// 按最后上报时间排序
		index.Fields("last_reported_at").
			Annotations(entsql.WithComments(true)),

		// 按最后心跳时间排序
		index.Fields("last_heartbeat_at").
			Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
