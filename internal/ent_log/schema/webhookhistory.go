package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WebhookHistory holds the schema definition for the WebhookHistory entity.
type WebhookHistory struct {
	ent.Schema
}

// Annotations of the WebhookHistory.
func (WebhookHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webhook_history"},
		schema.Comment("通知历史记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the WebhookHistory.
func (WebhookHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the WebhookHistory.
func (WebhookHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("通知日志序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("webhook_id").
			Immutable().
			Comment("通知序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("subscription_id").
			Optional().
			Immutable().
			Comment("通知订阅序号（可选，直接触发时为空）").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Comment("发送的URL").
			Annotations(entsql.WithComments(true)),

		field.String("title").
			Optional().
			Comment("发送的标题").
			Annotations(entsql.WithComments(true)),

		field.String("content").
			Optional().
			Comment("发送的内容").
			Annotations(entsql.WithComments(true)),

		field.Int("status_code").
			Optional().
			Comment("HTTP响应状态码").
			Annotations(entsql.WithComments(true)),

		field.Text("response").
			Optional().
			Comment("响应内容").
			Annotations(entsql.WithComments(true)),

		field.Bool("success").
			Default(false).
			Comment("是否发送成功").
			Annotations(entsql.WithComments(true)),

		field.String("error_message").
			Optional().
			Comment("错误信息").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the WebhookHistory.
func (WebhookHistory) Edges() []ent.Edge {
	return nil
}

// Indexes of the WebhookHistory.
func (WebhookHistory) Indexes() []ent.Index {
	return []ent.Index{
		// 按通知分组
		index.Fields("webhook_id").
			Annotations(entsql.WithComments(true)),

		// 按通知订阅分组
		index.Fields("subscription_id").
			Annotations(entsql.WithComments(true)),

		// 按发送状态分组
		index.Fields("success").
			Annotations(entsql.WithComments(true)),

		// 按发送时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 复合索引：通知+时间，用于查询特定通知的发送历史
		index.Fields("webhook_id", "created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 复合索引：状态+时间，用于查询失败记录
		index.Fields("success", "created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
