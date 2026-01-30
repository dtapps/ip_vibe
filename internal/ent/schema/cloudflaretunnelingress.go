package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CloudflareTunnelIngress holds the schema definition for the CloudflareTunnelIngress entity.
type CloudflareTunnelIngress struct {
	ent.Schema
}

// Annotations of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cloudflare_tunnel_ingress"},
		schema.Comment("Cloudflare 隧道入站规则"),
		entsql.WithComments(true),
	}
}

// Mixin of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("Cloudflare 隧道入站规则序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("cloudflare_tunnel_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 Cloudflare 隧道序号").
			Annotations(entsql.WithComments(true)),

		field.String("hostname").
			Optional().
			Comment("主机名").
			Annotations(entsql.WithComments(true)),

		field.String("service").
			Comment("服务").
			Annotations(entsql.WithComments(true)),

		field.String("path").
			Optional().
			Comment("路径").
			Annotations(entsql.WithComments(true)),

		field.JSON("origin_request", json.RawMessage{}).
			Optional().
			Comment("入站请求配置").
			Annotations(entsql.WithComments(true)),

		field.Time("sync_at").
			Optional().
			Comment("最近一次同步时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 CloudflareTunnel -> CloudflareTunnelIngress
		edge.From("cloudflare_tunnel", CloudflareTunnel.Type).
			Ref("ingresses").
			Field("cloudflare_tunnel_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 Cloudflare 隧道").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按隧道ID排序
		index.Fields("cloudflare_tunnel_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CloudflareTunnelIngress.
func (CloudflareTunnelIngress) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.CloudflareTunnelIngressLogHook(),
	}
}
