package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/ent/schema/schemavalidate"

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
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("CDN记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Optional().
			Comment("平台侧记录ID").
			Annotations(entsql.WithComments(true)),

		field.String("sub_domain").
			Validate(schemavalidate.ValidateSubDomain).
			Immutable().
			Comment("子域名（例如 @, www, blog, api）").
			Annotations(entsql.WithComments(true)),

		field.String("domain_name").
			// Optional(). // 先兼容旧数据
			Validate(schemavalidate.ValidateDomainName).
			Immutable(). // 先兼容旧数据
			Comment("域名名称（例如 example.com）").
			Annotations(entsql.WithComments(true)),

		field.String("full_domain").
			// Optional(). // 先兼容旧数据
			Validate(schemavalidate.ValidateFullDomain).
			Immutable(). // 先兼容旧数据
			Comment("完整域名（例如 www.example.com）").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetCDNRecordTypes...).
			Comment("记录类型（A、AAAA）").
			Annotations(entsql.WithComments(true)),

		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("CDN记录配置（JSON）").
			Annotations(entsql.WithComments(true)),

		field.String("cdn_domain").
			Optional().
			Comment("CDN边缘域名（平台提供）").
			Annotations(entsql.WithComments(true)),

		field.Int64("ip_group_id").
			Optional().
			Comment("IP组序号（二选一）").
			Annotations(entsql.WithComments(true)),

		field.Int64("device_group_id").
			Optional().
			Comment("设备组序号（二选一）").
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
			Sensitive(). // 敏感
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
			Comment("关联的IP组（二选一）").
			Annotations(entsql.WithComments(true)),

		// 一对多 DeviceGroup -> CDNRecord
		edge.From("device_group", DeviceGroup.Type).
			Ref("cdn_records").
			Field("device_group_id").
			Unique().
			Comment("关联的设备组（二选一）").
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
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按域名排序
		index.Fields("sub_domain").
			Annotations(entsql.WithComments(true)),

		// 按记录类型排序
		index.Fields("type").
			Annotations(entsql.WithComments(true)),

		// 按IP组序号排序
		index.Fields("ip_group_id").
			Annotations(entsql.WithComments(true)),

		// 按设备组序号排序
		index.Fields("device_group_id").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CDNRecord.
func (CDNRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.CDNRecordLogHook(),
	}
}
