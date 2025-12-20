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

// Mixin of the CDNRecord.
func (CDNRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the CDNRecord.
func (CDNRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("CDN记录序号").
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
			Values(schematype.GetCDNRecordTypes...).
			Comment("记录类型（A、AAAA）").
			Annotations(entsql.WithComments(true)),

		field.Enum("origin_protocol").
			Default(schematype.OriginProtocolFollow).
			Values(schematype.GetOriginProtocols...).
			Comment("回源协议（http、https、follow）").
			Annotations(entsql.WithComments(true)),

		field.Int("http_port").
			Default(80).
			Comment("HTTP 回源端口").
			Annotations(entsql.WithComments(true)),

		field.Int("https_port").
			Default(443).
			Comment("HTTPS 回源端口").
			Annotations(entsql.WithComments(true)),

		field.String("cdn_domain").
			Optional().
			Comment("CDN边缘域名（平台提供）").
			Annotations(entsql.WithComments(true)),

		field.Int64("ip_group_id").
			Comment("IP组序号").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Time("last_refresh_at").
			Optional().
			Comment("最后一次刷新成功时间").
			Annotations(entsql.WithComments(true)),

		field.String("last_api_response").
			Optional().
			Comment("最近一次接口的反馈").
			Annotations(entsql.WithComments(true)),

		field.Time("last_api_response_at").
			Optional().
			Comment("最近一次接口的反馈时间").
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
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 IpGroup -> CDNRecord
		edge.From("ip_group", IpGroup.Type).
			Ref("cdn_records").
			Field("ip_group_id").
			Unique().
			Required().
			Comment("关联的IP组").
			Annotations(entsql.WithComments(true)),

		// 一对一 CDNRecord -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("关联的DNS记录").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the CDNRecord.
func (CDNRecord) Indexes() []ent.Index {
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
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按IP组序号排序
		index.Fields("ip_group_id").
			Annotations(entsql.WithComments(true)),
	}
}
