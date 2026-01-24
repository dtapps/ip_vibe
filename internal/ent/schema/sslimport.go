package schema

import (
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// SSLImport holds the schema definition for the SSLImport entity.
type SSLImport struct {
	ent.Schema
}

// Annotations of the SSLImport.
func (SSLImport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ssl_import"},
		schema.Comment("SSL导入"),
		entsql.WithComments(true),
	}
}

// Mixin of the SSLImport.
func (SSLImport) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the SSLImport.
func (SSLImport) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("SSL导入序号").
			Annotations(entsql.WithComments(true)),

		field.Enum("source").
			Values(schematype.GetSSLImportSources...).
			Default(schematype.SSLImportSourceUnknown).
			Immutable().
			Comment("来源").
			Annotations(entsql.WithComments(true)),

		field.String("domain").
			Optional().
			Immutable().
			Comment("域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）").
			Annotations(entsql.WithComments(true)),

		field.Text("certificate").
			Optional().
			Immutable().
			Comment("证书(PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.Text("private_key").
			Optional().
			Sensitive(). // 敏感
			Immutable().
			Comment("私钥(PEM格式)").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the SSLImport.
func (SSLImport) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the SSLImport.
func (SSLImport) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按域名排序
		index.Fields("domain").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the SSLImport.
func (SSLImport) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.SSLImportLogHook(),
	}
}
