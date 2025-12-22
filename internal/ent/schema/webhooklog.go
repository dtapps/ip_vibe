package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WebhookLog holds the schema definition for the WebhookLog entity.
type WebhookLog struct {
	ent.Schema
}

// Annotations of the WebhookLog.
func (WebhookLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webhook_log"},
		schema.Comment("通知发送日志"),
		entsql.WithComments(true),
	}
}

// Mixin of the WebhookLog.
func (WebhookLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the WebhookLog.
func (WebhookLog) Fields() []ent.Field {
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

// Edges of the WebhookLog.
func (WebhookLog) Edges() []ent.Edge {
	return []ent.Edge{
		// 多对一：WebhookLog -> Webhook
		edge.From("webhook", Webhook.Type).
			Ref("logs").
			Field("webhook_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的通知（必需）").
			Annotations(entsql.WithComments(true)),

		// 多对一：WebhookLog -> WebhookSubscription
		edge.From("subscription", WebhookSubscription.Type).
			Ref("logs").
			Field("subscription_id").
			Unique().
			Immutable().
			Comment("关联的通知订阅（可选）").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the WebhookLog.
func (WebhookLog) Indexes() []ent.Index {
	return []ent.Index{
		// 按通知分组
		index.Fields("webhook_id").
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
