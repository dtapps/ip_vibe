package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// OperationLog holds the schema definition for the OperationLog entity.
type OperationLog struct {
	ent.Schema
}

// Annotations of the OperationLog.
func (OperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "operation_log"},
		schema.Comment("操作日志"),
		entsql.WithComments(true),
	}
}

// Fields of the OperationLog.
func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("记录序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Enum("op_type").
			Values(
				"data_create",
				"data_update",
				"data_delete",
				"device_report",
				"ip_refresh",
			).
			Immutable().
			Comment("操作类型").
			Annotations(entsql.WithComments(true)),

		field.Int64("obj_id").
			Immutable().
			Comment("对象主键").
			Annotations(entsql.WithComments(true)),

		field.Enum("obj_type").
			Values(
				"platform_config",
				"cdn_record",
				"dns_record",
				"device",
				"device_ip_report",
				"ip_group",
				"ip_group_device",
			).
			Immutable().
			Comment("对象类型").
			Annotations(entsql.WithComments(true)),

		field.String("old_value").
			Optional().
			Immutable().
			Comment("旧值").
			Annotations(entsql.WithComments(true)),

		field.String("new_value").
			Optional().
			Immutable().
			Comment("新值").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values("success", "failed").
			Immutable().
			Comment("状态").
			Annotations(entsql.WithComments(true)),

		field.String("message").
			Immutable().
			Comment("详情").
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the OperationLog.
func (OperationLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the OperationLog.
func (OperationLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("op_type").
			Annotations(entsql.WithComments(true)),

		index.Fields("obj_type").
			Annotations(entsql.WithComments(true)),
	}
}
