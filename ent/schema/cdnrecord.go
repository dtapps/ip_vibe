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

// CDNRecord holds the schema definition for the CDNRecord entity.
type CDNRecord struct {
	ent.Schema
}

// Annotations of the CDNRecord.
func (CDNRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cdn_record"},
		schema.Comment("CDN记录"),
		entsql.WithComments(true),
	}
}

// Fields of the CDNRecord.
func (CDNRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("CDN 记录序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Int64("config_id").
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Optional().
			Comment("平台侧记录 ID").
			Annotations(entsql.WithComments(true)),

		field.String("domain").
			Immutable().
			Comment("域名（如 www.example.com）").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values("A", "AAAA").
			Immutable().
			Comment("记录类型（A、AAAA等）").
			Annotations(entsql.WithComments(true)),

		field.Enum("origin_protocol").
			Default("follow").
			Values("http", "https", "follow").
			Immutable().
			Comment("回源协议").
			Annotations(entsql.WithComments(true)),

		field.Int("http_port").
			Default(80).
			Immutable().
			Comment("HTTP 回源端口").
			Annotations(entsql.WithComments(true)),

		field.Int("https_port").
			Default(443).
			Immutable().
			Comment("HTTPS 回源端口").
			Annotations(entsql.WithComments(true)),

		field.String("cdn_domain").
			Optional().
			Comment("CDN 边缘域名（平台提供）").
			Annotations(entsql.WithComments(true)),

		field.Int64("ip_group_id").
			Immutable().
			Comment("IP 组序号").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Time("last_refresh_at").
			Optional().
			Comment("最后一次刷新成功时间").
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

// Edges of the CDNRecord.
func (CDNRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> CDNRecord
		edge.From("config", PlatformConfig.Type).
			Ref("cdn_records").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 IpGroup -> CDNRecord
		edge.From("ip_group", IpGroup.Type).
			Ref("cdn_records").
			Field("ip_group_id").
			Unique().
			Required().
			Immutable().
			Comment("IP 组").
			Annotations(entsql.WithComments(true)),

		// 一对一 CDNRecord -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("DNS 记录").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the CDNRecord.
func (CDNRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		index.Fields("type").
			Annotations(entsql.WithComments(true)),

		index.Fields("ip_group_id").
			Annotations(entsql.WithComments(true)),

		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),
	}
}
