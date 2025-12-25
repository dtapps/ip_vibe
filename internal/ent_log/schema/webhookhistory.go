package schema

import (
	"encoding/json"
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
		schemamixin.AtLogMixin{},
	}
}

// Fields of the WebhookHistory.
func (WebhookHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("通知历史记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("webhook_id").
			Immutable().
			Comment("关联 Webhook 序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("subscription_id").
			Optional().
			Immutable().
			Comment("关联 WebhookSubscription 序号").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Immutable().
			Comment("发送的URL").
			Annotations(entsql.WithComments(true)),

		field.String("title").
			Optional().
			Immutable().
			Comment("发送的标题").
			Annotations(entsql.WithComments(true)),

		field.String("content").
			Optional().
			Immutable().
			Comment("发送的内容").
			Annotations(entsql.WithComments(true)),

		field.JSON("response", json.RawMessage{}).
			Optional().
			Immutable().
			Comment("响应内容").
			Annotations(entsql.WithComments(true)),

		field.Bool("success").
			Default(false).
			Immutable().
			Comment("是否发送成功").
			Annotations(entsql.WithComments(true)),

		field.String("error_message").
			Optional().
			Immutable().
			Comment("错误信息").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the WebhookHistory.
func (WebhookHistory) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the WebhookHistory.
func (WebhookHistory) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按关联 Webhook ID分组
		index.Fields("webhook_id").
			Annotations(entsql.WithComments(true)),

		// 按关联 WebhookSubscription ID分组
		index.Fields("subscription_id").
			Annotations(entsql.WithComments(true)),

		// 按发送状态分组
		index.Fields("success").
			Annotations(entsql.WithComments(true)),

		// 复合索引：通知+时间，用于查询特定通知的发送历史
		index.Fields("webhook_id", "created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 复合索引：状态+时间，用于查询失败记录
		index.Fields("success", "created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
