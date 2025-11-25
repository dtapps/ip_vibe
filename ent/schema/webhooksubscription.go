package schema

import (
	"vibe_ip/ent/schema/schemamixin"
	"vibe_ip/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WebhookSubscription holds the schema definition for the WebhookSubscription entity.
type WebhookSubscription struct {
	ent.Schema
}

// Annotations of the WebhookSubscription.
func (WebhookSubscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webhook_subscription"},
		schema.Comment("通知订阅 (多对多关联)"),
		entsql.WithComments(true),
	}
}

// Mixin of the WebhookSubscription.
func (WebhookSubscription) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the WebhookSubscription.
func (WebhookSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("通知订阅序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("webhook_id").
			Immutable().
			Comment("通知序号").
			Annotations(entsql.WithComments(true)),

		field.Enum("obj_type").
			Values(schematype.GetWebhookSubscriptionObjType()...).
			Immutable().
			Comment("订阅对象类型").
			Annotations(entsql.WithComments(true)),

		field.Int64("obj_id").
			Immutable().
			Comment("订阅对象主键").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the WebhookSubscription.
func (WebhookSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("webhook", Webhook.Type).
			Field("webhook_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的通知").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the WebhookSubscription.
func (WebhookSubscription) Indexes() []ent.Index {
	return []ent.Index{
		// 确保一个 Webhook 不能对同一个对象（类型+ID）重复订阅
		index.Fields("webhook_id", "obj_type", "obj_id").
			Unique(),
	}
}
