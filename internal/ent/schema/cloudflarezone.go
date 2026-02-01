package schema

import (
	"encoding/json"
	// "vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CloudflareZone holds the schema definition for the CloudflareZone entity.
type CloudflareZone struct {
	ent.Schema
}

// Annotations of the CloudflareZone.
func (CloudflareZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cloudflare_zone"},
		schema.Comment("Cloudflare 站点"),
		entsql.WithComments(true),
	}
}

// Mixin of the CloudflareZone.
func (CloudflareZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the CloudflareZone.
func (CloudflareZone) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("Cloudflare 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("zone_id").
			Immutable().
			Comment("站点 ID").
			Annotations(entsql.WithComments(true)),

		field.JSON("account", json.RawMessage{}).
			Optional().
			Comment("区域所属的帐户").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Optional().
			Comment("域名").
			Annotations(entsql.WithComments(true)),

		field.Strings("name_servers").
			Optional().
			Comment("Cloudflare分配给区域的名称服务器").
			Annotations(entsql.WithComments(true)),

		field.String("original_dnshost").
			Optional().
			Comment("切换到Cloudflare时的DNS主机").
			Annotations(entsql.WithComments(true)),

		field.Strings("original_name_servers").
			Optional().
			Comment("迁移到Cloudflare之前的原始名称服务器").
			Annotations(entsql.WithComments(true)),

		field.String("original_registrar").
			Optional().
			Comment("转换到Cloudflare时的域名注册商").
			Annotations(entsql.WithComments(true)),

		field.Bool("paused").
			Optional().
			Comment("指示区域是否仅使用Cloudflare DNS服务。 true值意味着该区域将不会获得安全或性能方面的好处").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetCloudflareZoneStatuss...).
			Optional().
			Comment("Cloudflare上的区域状态").
			Annotations(entsql.WithComments(true)),

		field.String("type").
			Optional().
			Comment("完整区域意味着DNS托管在Cloudflare上。部分区域通常是合作伙伴托管的区域或CNAME设置").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CloudflareZone.
func (CloudflareZone) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> CloudflareZone
		edge.From("config", PlatformConfig.Type).
			Ref("cloudflare_zones").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 CloudflareZone -> CloudflareZoneRecord
		edge.To("records", CloudflareZoneRecord.Type).
			Comment("关联的Cloudflare站点记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 CloudflareZone 时，其所有 CloudflareZoneRecord 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the CloudflareZone.
func (CloudflareZone) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("zone_id").
			Annotations(entsql.WithComments(true)),

		// 按平台配置序号+站点ID排序
		index.Fields("config_id", "zone_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CloudflareZone.
func (CloudflareZone) Hooks() []ent.Hook {
	return []ent.Hook{
		// schemahook.CloudflareZoneLogHook(),
	}
}
