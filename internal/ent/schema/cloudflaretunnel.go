package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CloudflareTunnel holds the schema definition for the CloudflareTunnel entity.
type CloudflareTunnel struct {
	ent.Schema
}

// Annotations of the CloudflareTunnel.
func (CloudflareTunnel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cloudflare_tunnel"},
		schema.Comment("Cloudflare 隧道"),
		entsql.WithComments(true),
	}
}

// Mixin of the CloudflareTunnel.
func (CloudflareTunnel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the CloudflareTunnel.
func (CloudflareTunnel) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("Cloudflare 隧道序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("tunnel_id").
			Comment("Cloudflare 隧道ID").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("隧道名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetCloudflareTunnelStatuss...).
			Comment("隧道状态").
			Annotations(entsql.WithComments(true)),

		field.JSON("connections", json.RawMessage{}).
			Optional().
			Comment("隧道连接配置（JSON）").
			Annotations(entsql.WithComments(true)),

		field.Time("conns_active_at").
			Optional().
			Comment("隧道连接活跃时间").
			Annotations(entsql.WithComments(true)),

		field.Time("remote_created_at").
			Immutable().
			Comment("隧道创建时间").
			Annotations(entsql.WithComments(true)),

		field.Time("sync_at").
			Optional().
			Comment("最近一次同步时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CloudflareTunnel.
func (CloudflareTunnel) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> CloudflareTunnel
		edge.From("config", PlatformConfig.Type).
			Ref("cloudflare_tunnels").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 CloudflareTunnel -> CloudflareTunnelIngress
		edge.To("ingresses", CloudflareTunnelIngress.Type).
			Comment("关联的Cloudflare隧道入站规则").
			Annotations(entsql.Annotation{
				// 级联删除：删除 CloudflareTunnel 时，其所有 CloudflareTunnelIngress 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the CloudflareTunnel.
func (CloudflareTunnel) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按隧道ID排序
		index.Fields("tunnel_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CloudflareTunnel.
func (CloudflareTunnel) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.CloudflareTunnelLogHook(),
	}
}
