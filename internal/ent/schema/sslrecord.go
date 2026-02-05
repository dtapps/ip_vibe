package schema

import (
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

// SSLRecord holds the schema definition for the SSLRecord entity.
type SSLRecord struct {
	ent.Schema
}

// Annotations of the SSLRecord.
func (SSLRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ssl_record"},
		schema.Comment("SSL记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the SSLRecord.
func (SSLRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the SSLRecord.
func (SSLRecord) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("SSL记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.String("domain").
			NotEmpty().
			Immutable().
			Comment("域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）").
			Annotations(entsql.WithComments(true)),

		field.String("cert_domain").
			Optional().
			Comment("证书域名").
			Annotations(entsql.WithComments(true)),

		field.String("cert_url").
			Optional().
			Comment("证书URL").
			Annotations(entsql.WithComments(true)),

		field.String("cert_stable_url").
			Optional().
			Comment("稳定证书URL").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_private_key").
			Optional().
			Sensitive(). // 敏感
			Comment("私钥(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_certificate").
			Optional().
			Comment("证书(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_issuer_certificate").
			Optional().
			Comment("颁发者证书(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_csr").
			Optional().
			Comment("证书签名请求(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Time("not_before").
			Optional().
			Comment("证书生效时间").
			Annotations(entsql.WithComments(true)),

		field.Time("not_after").
			Optional().
			Comment("证书过期时间").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetSSLRecordStatuss...).
			Default(schematype.SSLRecordStatusUnknown).
			Comment("证书状态").
			Annotations(entsql.WithComments(true)),

		field.Enum("provider").
			Values(schematype.GetSSLRecordProviders...).
			Default(schematype.SSLRecordProviderUnknown).
			Comment("CA提供商").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Int("renewal_days").
			Optional().
			Default(30).
			Comment("过期前多少天开始续期").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_auto_renewal").
			Optional().
			Default(false).
			Comment("是否自动续期").
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

// Edges of the SSLRecord.
func (SSLRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> SSLRecord
		edge.From("config", PlatformConfig.Type).
			Ref("ssl_records").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the SSLRecord.
func (SSLRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按域名排序
		index.Fields("domain").
			Annotations(entsql.WithComments(true)),

		// 按证书状态排序
		index.Fields("status").
			Annotations(entsql.WithComments(true)),

		// 按CA提供商排序
		index.Fields("provider").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),

		// 按是否自动续期排序
		index.Fields("is_auto_renewal").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the SSLRecord.
func (SSLRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.SSLRecordLogHook(),
	}
}
