package schema

import (
	"vibe_ip/ent/schema/schemamixin"
	"vibe_ip/ent/schema/schematype"

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

// Mixin of the OperationLog.
func (OperationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the OperationLog.
func (OperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("日志序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.Enum("op_type").
			Values(schematype.GetLogOpType()...).
			Immutable().
			Comment("操作类型").
			Annotations(entsql.WithComments(true)),

		field.Int64("obj_id").
			Immutable().
			Comment("对象主键").
			Annotations(entsql.WithComments(true)),

		field.Enum("obj_type").
			Values(schematype.GetLogObjType()...).
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
			Values(schematype.GetLogStatus()...).
			Immutable().
			Comment("状态").
			Annotations(entsql.WithComments(true)),

		field.String("message").
			Immutable().
			Comment("详情").
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
		// 按操作类型排序
		index.Fields("op_type").
			Annotations(entsql.WithComments(true)),

		// 按对象主键排序
		index.Fields("obj_id").
			Annotations(entsql.WithComments(true)),

		// 按对象类型排序
		index.Fields("obj_type").
			Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.WithComments(true)),
	}
}
