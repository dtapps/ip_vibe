package schema

import (
	"vibe_ip/ent/schema/schemamixin"
	"vibe_ip/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Webhook holds the schema definition for the Webhook entity.
type Webhook struct {
	ent.Schema
}

// Annotations of the Webhook.
func (Webhook) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webhook"},
		schema.Comment("通知"),
		entsql.WithComments(true),
	}
}

// Mixin of the Webhook.
func (Webhook) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the Webhook.
func (Webhook) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("通知序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.String("name").
			Comment("通知名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("platform").
			Values(schematype.GetWebhookPlatform()...).
			Comment("通知平台类型").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Comment("Webhook URL / API Key").
			Annotations(entsql.WithComments(true)),

		// 密钥：用于签名验证或特殊 Token
		field.String("secret").
			Optional().
			Comment("Webhook 签名密钥或额外Token").
			Annotations(entsql.WithComments(true)),

		field.JSON("events", []string{}).
			Comment("订阅的事件类型列表").
			Annotations(entsql.WithComments(true)),

		field.Text("template_title").
			Optional().
			Comment("通知标题 Go Template").
			Annotations(entsql.WithComments(true)),

		field.Text("template_content").
			Optional().
			Comment("通知内容 Go Template").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否启用").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the Webhook.
func (Webhook) Edges() []ent.Edge {
	return nil
}
