package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// IpGroupDevice holds the schema definition for the IpGroupDevice entity.
type IpGroupDevice struct {
	ent.Schema
}

// Annotations of the IpGroupDevice.
func (IpGroupDevice) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ip_group_device"},
		schema.Comment("IP组设备"),
		entsql.WithComments(true),
		field.ID("device_id", "ip_group_id"),
	}
}

// Fields of the IpGroupDevice.
func (IpGroupDevice) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("IP组设备序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Int64("device_id").
			Comment("设备序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("ip_group_id").
			Comment("IP组序号").
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the IpGroupDevice.
func (IpGroupDevice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("device", Device.Type).
			Unique().
			Required().
			Field("device_id"),
		edge.To("ip_group", IpGroup.Type).
			Unique().
			Required().
			Field("ip_group_id"),
	}
}

// Indexes of the IpGroupDevice.
func (IpGroupDevice) Indexes() []ent.Index {
	return []ent.Index{}
}
