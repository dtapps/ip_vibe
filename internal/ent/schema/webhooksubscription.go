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
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("通知订阅序号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("通知订阅名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("通知订阅描述").
			Annotations(entsql.WithComments(true)),

		field.Int64("webhook_id").
			Positive(). // 正数
			Immutable().
			Comment("通知序号").
			Annotations(entsql.WithComments(true)),

		field.Enum("obj_type").
			Values(schematype.GetWebhookSubscriptionObjTypes...).
			Immutable().
			Comment("对象类型").
			Annotations(entsql.WithComments(true)),

		field.Int64("obj_id").
			Positive(). // 正数
			Immutable().
			Comment("对象主键").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
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
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按通知序号排序
		index.Fields("webhook_id").
			Annotations(entsql.WithComments(true)),

		// 按对象类型排序
		index.Fields("obj_type").
			Annotations(entsql.WithComments(true)),

		// 按对象主键排序
		index.Fields("obj_id").
			Annotations(entsql.WithComments(true)),

		// 按是否活跃排序
		index.Fields("is_active").
			Annotations(entsql.WithComments(true)),

		// 按对象类型和对象主键排序
		index.Fields("obj_type", "obj_id").
			Annotations(entsql.WithComments(true)),

		// 按对象类型和对象主键和创建时间排序
		index.Fields("obj_type", "obj_id", "created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 确保一个通知不能对同一个对象（类型+ID）重复订阅
		index.Fields("webhook_id", "obj_type", "obj_id").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}
