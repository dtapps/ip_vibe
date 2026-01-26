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

// DeviceGroupDevice holds the schema definition for the DeviceGroupDevice entity.
type DeviceGroupDevice struct {
	ent.Schema
}

// Annotations of the DeviceGroupDevice.
func (DeviceGroupDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_group_device"},
		schema.Comment("设备组设备 (多对多关联)"),
		entsql.WithComments(true),
		field.ID("device_id", "device_group_id"),
	}
}

// Mixin of the DeviceGroupDevice.
func (DeviceGroupDevice) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the DeviceGroupDevice.
func (DeviceGroupDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("设备组设备序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("device_id").
			Positive(). // 正数
			Comment("设备序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("device_group_id").
			Positive(). // 正数
			Comment("设备组序号").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DeviceGroupDevice.
func (DeviceGroupDevice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("device", Device.Type).
			Unique().
			Required().
			Field("device_id").
			Comment("关联的设备").
			Annotations(entsql.WithComments(true)),

		edge.To("device_group", DeviceGroup.Type).
			Unique().
			Required().
			Field("device_group_id").
			Comment("关联的设备组").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DeviceGroupDevice.
func (DeviceGroupDevice) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按设备序号排序
		index.Fields("device_id").
			Annotations(entsql.WithComments(true)),

		// 按设备组序号排序
		index.Fields("device_group_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the DeviceGroupDevice.
func (DeviceGroupDevice) Hooks() []ent.Hook {
	return []ent.Hook{}
}
