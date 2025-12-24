package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SSLCertificateHistory holds the schema definition for the SSLCertificateHistory entity.
type SSLCertificateHistory struct {
	ent.Schema
}

// Annotations of the SSLCertificateHistory.
func (SSLCertificateHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ssl_certificate_history"},
		schema.Comment("SSL证书申请历史记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the SSLCertificateHistory.
func (SSLCertificateHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the SSLCertificateHistory.
func (SSLCertificateHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("SSL证书申请历史记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("ssl_record_id").
			Immutable().
			Comment("关联 SSLRecord ID").
			Annotations(entsql.WithComments(true)),

		field.String("cert_domain").
			Optional().
			Immutable().
			Comment("证书域名").
			Annotations(entsql.WithComments(true)),

		field.String("cert_url").
			Optional().
			Immutable().
			Comment("证书URL").
			Annotations(entsql.WithComments(true)),

		field.String("cert_stable_url").
			Optional().
			Immutable().
			Comment("稳定证书URL").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_private_key").
			Sensitive(). // 敏感
			Immutable().
			Comment("私钥(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_certificate").
			Immutable().
			Comment("证书(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_issuer_certificate").
			Optional().
			Immutable().
			Comment("颁发者证书(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("cert_csr").
			Optional().
			Immutable().
			Comment("证书签名请求(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Time("not_before").
			Optional().
			Immutable().
			Comment("证书生效时间").
			Annotations(entsql.WithComments(true)),

		field.Time("not_after").
			Optional().
			Immutable().
			Comment("证书过期时间").
			Annotations(entsql.WithComments(true)),

		field.Enum("provider").
			Values(schematype.GetSSLRecordProviders...).
			Default(schematype.SSLRecordProviderLetsencrypt).
			Immutable().
			Comment("CA提供商").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the SSLCertificateHistory.
func (SSLCertificateHistory) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the SSLCertificateHistory.
func (SSLCertificateHistory) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按关联 SSLRecord ID 分组
		index.Fields("ssl_record_id").
			Annotations(entsql.WithComments(true)),
	}
}
