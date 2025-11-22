package schema

import (
	"encoding/json"
	"time"

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

// Fields of the PlatformConfig.
func (PlatformConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("配置序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.String("name").
			Comment("配置名称（便于展示）").
			Annotations(entsql.WithComments(true)),

		field.Enum("platform").
			Values("cloudflare", "edgeone", "dnspod").
			Immutable().
			Comment("平台类型（edgeone / cloudflare / dnspod …）").
			Annotations(entsql.WithComments(true)),

		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("平台密钥等敏感信息（JSON）").
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

// Edges of the PlatformConfig.
func (PlatformConfig) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> DNSRecord
		edge.To("dns_records", DNSRecord.Type).
			Comment("DNS记录").
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 一对多 PlatformConfig -> CDNRecord
		edge.To("cdn_records", CDNRecord.Type).
			Comment("CDN记录").
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the PlatformConfig.
func (PlatformConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform").
			Annotations(entsql.WithComments(true)),
	}
}
