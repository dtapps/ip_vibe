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

// CloudflareZoneRecord holds the schema definition for the CloudflareZoneRecord entity.
type CloudflareZoneRecord struct {
	ent.Schema
}

// Annotations of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cloudflare_zone_record"},
		schema.Comment("Cloudflare 站点记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("Cloudflare 站点记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("cloudflare_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 Cloudflare 站点序号").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Immutable().
			Comment("记录 ID").
			Annotations(entsql.WithComments(true)),

		field.String("comment").
			Optional().
			Comment("有关DNS记录的注释或说明。 此字段对DNS响应没有影响。").
			Annotations(entsql.WithComments(true)),

		field.String("content").
			Optional().
			Comment("单个Base64编码的OpenPGP可传输公钥（RFC 4880第11.1节）").
			Annotations(entsql.WithComments(true)),

		field.Bool("proxiable").
			Optional().
			Comment("Cloudflare是否可以代理该记录").
			Annotations(entsql.WithComments(true)),

		field.Bool("proxied").
			Optional().
			Comment("记录是否获得Cloudflare的性能和安全优势").
			Annotations(entsql.WithComments(true)),

		field.JSON("settings", json.RawMessage{}).
			Optional().
			Comment("DNS记录的设置").
			Annotations(entsql.WithComments(true)),

		field.Float("ttl").
			Optional().
			Comment("DNS记录的生存时间（TTL），以秒为单位。 设置为1表示“自动”。 值必须介于60和86400之间，对于企业区域，最小值减少到30").
			Annotations(entsql.WithComments(true)),

		field.Enum("type").
			Values(schematype.GetCloudflareZoneRecordTypes...).
			Optional().
			Comment("数据录类型").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 CloudflareZone -> CloudflareZoneRecord
		edge.From("cloudflare_zone", CloudflareZone.Type).
			Ref("records").
			Field("cloudflare_zone_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 Cloudflare 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("cloudflare_zone_id").
			Annotations(entsql.WithComments(true)),

		// 按记录ID排序
		index.Fields("record_id").
			Annotations(entsql.WithComments(true)),

		// 按站点ID+记录ID排序
		index.Fields("cloudflare_zone_id", "record_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CloudflareZoneRecord.
func (CloudflareZoneRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		// schemahook.CloudflareZoneRecordLogHook(),
	}
}
