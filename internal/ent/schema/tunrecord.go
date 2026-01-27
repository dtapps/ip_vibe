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

// TUNRecord holds the schema definition for the TUNRecord entity.
type TUNRecord struct {
	ent.Schema
}

// Annotations of the TUNRecord.
func (TUNRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tun_record"},
		schema.Comment("TUN记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the TUNRecord.
func (TUNRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the TUNRecord.
func (TUNRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("TUN记录序号").
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
			Comment("子域名（如 www, blog, api）").
			Annotations(entsql.WithComments(true)),

		field.String("domain_path").
			Optional().
			Comment("域名路径（如 /api/v1）").
			Annotations(entsql.WithComments(true)),

		field.Enum("serve_type").
			Values(schematype.GetTUNRecordServeTypes...).
			Comment("服务类型").
			Annotations(entsql.WithComments(true)),

		field.String("serve_uri").
			Optional().
			Comment("服务URI（如 127.0.0.1:8080）").
			Annotations(entsql.WithComments(true)),

		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("TUN记录配置（JSON）").
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

// Edges of the TUNRecord.
func (TUNRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> TUNRecord
		edge.From("config", PlatformConfig.Type).
			Ref("tun_records").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the TUNRecord.
func (TUNRecord) Indexes() []ent.Index {
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

		// 按服务类型排序
		index.Fields("serve_type").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the TUNRecord.
func (TUNRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.TUNRecordLogHook(),
	}
}
