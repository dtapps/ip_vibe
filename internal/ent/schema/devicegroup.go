package schema

import (
	"vibe_ip/internal/ent/schema/schemadefault"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DeviceGroup holds the schema definition for the DeviceGroup entity.
type DeviceGroup struct {
	ent.Schema
}

// Annotations of the DeviceGroup.
func (DeviceGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_group"},
		schema.Comment("设备组"),
		entsql.WithComments(true),
	}
}

// Mixin of the DeviceGroup.
func (DeviceGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the DeviceGroup.
func (DeviceGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("设备组序号").
			Annotations(entsql.WithComments(true)),

		field.String("code").
			DefaultFunc(schemadefault.DefaultCode).
			Immutable().
			Comment("设备组编码").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("设备组名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("设备组描述").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DeviceGroup.
func (DeviceGroup) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对多：Device <-> DeviceGroup
		edge.From("devices", Device.Type).
			Ref("device_groups").
			Through("device_group_device", DeviceGroupDevice.Type).
			Comment("关联的设备").
			Annotations(entsql.WithComments(true)),

		// 一对多 DeviceGroup -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("关联的DNS记录").
			Annotations(entsql.WithComments(true)),

		// 一对多 DeviceGroup -> CDNRecord
		edge.To("cdn_records", CDNRecord.Type).
			Comment("关联的CDN记录").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DeviceGroup.
func (DeviceGroup) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按设备组编码排序
		index.Fields("code").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the DeviceGroup.
func (DeviceGroup) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.DeviceGroupLogHook(),
	}
}
