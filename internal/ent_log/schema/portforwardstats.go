package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PortForwardStats holds the schema definition for the PortForwardStats entity.
type PortForwardStats struct {
	ent.Schema
}

// Annotations of the PortForwardStats.
func (PortForwardStats) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "port_forward_stats"},
		schema.Comment("端口转发统计快照记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the PortForwardStats.
func (PortForwardStats) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the PortForwardStats.
func (PortForwardStats) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("端口转发统计记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("device_port_forward_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 DevicePortForward 序号").
			Annotations(entsql.WithComments(true)),

		field.Time("start_time").
			Immutable().
			Comment("转发启动时间 (UTC)").
			Annotations(entsql.WithComments(true)),

		field.Int64("active_conn").
			Optional().
			Immutable().
			Comment("瞬时活跃连接数").
			Annotations(entsql.WithComments(true)),

		field.Int64("total_conn").
			Optional().
			Immutable().
			Comment("累计连接数").
			Annotations(entsql.WithComments(true)),

		field.Int64("bytes_in").
			Optional().
			Immutable().
			Comment("累计入站流量 (bytes)").
			Annotations(entsql.WithComments(true)),

		field.Int64("bytes_out").
			Optional().
			Immutable().
			Comment("累计出站流量 (bytes)").
			Annotations(entsql.WithComments(true)),

		field.Int64("udp_packets_in").
			Optional().
			Immutable().
			Comment("累计入站 UDP 包数").
			Annotations(entsql.WithComments(true)),

		field.Int64("udp_packets_out").
			Optional().
			Immutable().
			Comment("累计出站 UDP 包数").
			Annotations(entsql.WithComments(true)),

		field.Int64("errors").
			Optional().
			Immutable().
			Comment("累计错误数").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the PortForwardStats.
func (PortForwardStats) Edges() []ent.Edge {
	return nil
}

// Indexes of the PortForwardStats.
func (PortForwardStats) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间降序排序，便于查询最新记录
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按关联 DevicePortForward 序号排序，便于过滤记录
		index.Fields("device_port_forward_id").
			Annotations(entsql.WithComments(true)),

		// 复合索引：关联 DevicePortForward 序号 + 创建时间，用于高效时间范围查询
		index.Fields("device_port_forward_id", "created_at").
			Annotations(entsql.WithComments(true)),
	}
}
