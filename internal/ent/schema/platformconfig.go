package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PlatformConfig holds the schema definition for the PlatformConfig entity.
type PlatformConfig struct {
	ent.Schema
}

// Annotations of the PlatformConfig.
func (PlatformConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "platform_config"},
		schema.Comment("平台配置"),
		entsql.WithComments(true),
	}
}

// Mixin of the PlatformConfig.
func (PlatformConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the PlatformConfig.
func (PlatformConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Enum("capability").
			Default(schematype.ConfigCapabilityAll).
			Values(schematype.GetConfigCapabilitys...).
			Comment("平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("平台配置名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("平台配置描述").
			Annotations(entsql.WithComments(true)),

		field.Enum("platform").
			Values(schematype.GetConfigPlatforms...).
			Immutable().
			Comment("平台类型").
			Annotations(entsql.WithComments(true)),

		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("平台密钥等敏感信息（JSON）").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the PlatformConfig.
func (PlatformConfig) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("关联的DNS记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 PlatformConfig 时，其所有 DNSRecord 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 一对多 PlatformConfig -> CDNRecord
		edge.To("cdn_records", CDNRecord.Type).
			Comment("关联的CDN记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 PlatformConfig 时，其所有 CDNRecord 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the PlatformConfig.
func (PlatformConfig) Indexes() []ent.Index {
	return []ent.Index{
		// 按平台支持的服务能力排序
		index.Fields("capability").
			Annotations(entsql.WithComments(true)),

		// 按平台类型排序
		index.Fields("platform").
			Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
