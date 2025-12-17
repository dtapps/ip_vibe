package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DNSRecord holds the schema definition for the DNSRecord entity.
type DNSRecord struct {
	ent.Schema
}

// Annotations of the DNSRecord.
func (DNSRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "dns_record"},
		schema.Comment("DNS记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the DNSRecord.
func (DNSRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the DNSRecord.
func (DNSRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("DNS记录序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Int64("config_id").
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Optional().
			Comment("平台侧记录ID").
			Annotations(entsql.WithComments(true)),

		// field.String("domain").
		// 	Immutable().
		// 	Comment("域名（如 www.example.com）").
		// 	Annotations(entsql.WithComments(true)),

		field.String("sub_domain").
			Immutable().
			Comment("子域名（如 www, blog, api）").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetDNSRecordTypes...).
			Immutable().
			Comment("记录类型（A、AAAA、CNAME等）").
			Annotations(entsql.WithComments(true)),

		field.Int64("ip_group_id").
			Optional().
			Comment("IP组序号（与 cdn_record_id 二选一）").
			Annotations(entsql.WithComments(true)),

		field.Int64("cdn_record_id").
			Optional().
			Unique().
			Comment("CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Time("last_refresh_at").
			Optional().
			Comment("最后一次刷新成功时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DNSRecord.
func (DNSRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> DNSRecord
		edge.From("config", PlatformConfig.Type).
			Ref("dns_records").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 IpGroup -> DNSRecord
		edge.From("ip_group", IpGroup.Type).
			Ref("dns_records").
			Field("ip_group_id").
			Unique().
			Comment("关联的IP组（二选一）").
			Annotations(entsql.WithComments(true)),

		// 一对一 CDNRecord -> DNSRecord
		edge.From("cdn_record", CDNRecord.Type).
			Ref("dns_records").
			Field("cdn_record_id").
			Unique().
			Comment("关联的CDN记录（二选一）").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DNSRecord.
func (DNSRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按域名排序
		index.Fields("sub_domain").
			Annotations(entsql.WithComments(true)),

		// 按记录类型排序
		index.Fields("type").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),

		// 确保一个域名存在多个相同类型的记录
		// index.Fields("sub_domain", "type").
		// 	Unique().
		// 	Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按IP组序号排序
		index.Fields("ip_group_id").
			Annotations(entsql.WithComments(true)),

		// 按CDN记录序号排序
		index.Fields("cdn_record_id").
			Annotations(entsql.WithComments(true)),
	}
}
