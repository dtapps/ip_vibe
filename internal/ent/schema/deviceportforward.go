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

// DevicePortForward holds the schema definition for the DevicePortForward entity.
type DevicePortForward struct {
	ent.Schema
}

// Annotations of the DevicePortForward.
func (DevicePortForward) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_port_forward"},
		schema.Comment("设备端口转发"),
		entsql.WithComments(true),
	}
}

// Mixin of the DevicePortForward.
func (DevicePortForward) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the DevicePortForward.
func (DevicePortForward) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("设备端口转发序号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			NotEmpty().
			Comment("设备端口转发名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("设备端口转发描述").
			Annotations(entsql.WithComments(true)),

		field.Int("local_port").
			Positive(). // 正数
			Comment("本地监听端口").
			Annotations(entsql.WithComments(true)),

		field.String("remote_host").
			NotEmpty().
			Comment("远程目标地址").
			Annotations(entsql.WithComments(true)),

		field.Int("remote_port").
			Positive(). // 正数
			Comment("远程目标端口").
			Annotations(entsql.WithComments(true)),

		field.JSON("protocols", []string{}).
			Default([]string{}).
			Comment("协议列表，支持 tcp4, tcp6, udp4, udp6").
			Annotations(entsql.WithComments(true)),

		field.Bool("auto_forward").
			Default(false).
			Comment("防火墙自动放行").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DevicePortForward.
func (DevicePortForward) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("device", Device.Type).
			Ref("port_forwards").
			Unique().
			Required().
			Comment("关联的设备").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DevicePortForward.
func (DevicePortForward) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the DevicePortForward.
func (DevicePortForward) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.DevicePortForwardLogHook(),
	}
}
