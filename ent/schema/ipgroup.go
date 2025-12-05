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

// IpGroup holds the schema definition for the IpGroup entity.
type IpGroup struct {
	ent.Schema
}

// Annotations of the IpGroup.
func (IpGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ip_group"},
		schema.Comment("IP组"),
		entsql.WithComments(true),
	}
}

// Mixin of the IpGroup.
func (IpGroup) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the IpGroup.
func (IpGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("IP组序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.String("record_id").
			Optional().
			Comment("平台侧记录ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("IP组名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("IP组描述").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the IpGroup.
func (IpGroup) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对多：Device <-> IPGroup
		edge.From("devices", Device.Type).
			Ref("ip_groups").
			Through("ip_group_device", IpGroupDevice.Type).
			Comment("关联的设备").
			Annotations(entsql.WithComments(true)),

		// 一对多 IpGroup -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("关联的DNS记录").
			Annotations(entsql.WithComments(true)),

		// 一对多 IpGroup -> CDNRecord
		edge.To("cdn_records", CDNRecord.Type).
			Comment("关联的CDN记录").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the IpGroup.
func (IpGroup) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.WithComments(true)),
	}
}
