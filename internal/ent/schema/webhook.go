package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

		field.String("description").
			Optional().
			Comment("通知描述").
			Annotations(entsql.WithComments(true)),

		field.Enum("platform").
			Values(schematype.GetWebhookPlatforms...).
			Immutable().
			Comment("通知平台类型").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Comment("Webhook URL").
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
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the Webhook.
func (Webhook) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 Webhook -> WebhookSubscription
		edge.To("subscriptions", WebhookSubscription.Type).
			Comment("关联的通知订阅").
			Annotations(entsql.Annotation{
				// 级联删除：删除 Webhook 时，其所有 WebhookSubscription 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the Webhook.
func (Webhook) Indexes() []ent.Index {
	return []ent.Index{
		// 按通知平台类型排序
		index.Fields("platform").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.WithComments(true)),
	}
}
