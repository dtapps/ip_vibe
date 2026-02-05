package schema

import (
	"time"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CronJobLog holds the schema definition for the CronJobLog entity.
type CronJobLog struct {
	ent.Schema
}

// Annotations of the CronJobLog.
func (CronJobLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cron_job_log"},
		schema.Comment("定时任务日志"),
		entsql.WithComments(true),
	}
}

// Mixin of the CronJobLog.
func (CronJobLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the CronJobLog.
func (CronJobLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("任务日志序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("cron_job_id").
			Comment("关联 定时任务 序号").
			Annotations(entsql.WithComments(true)),

		field.String("handler_name").
			Comment("执行时的处理器名").
			Annotations(entsql.WithComments(true)),

		field.Time("started_at").
			Default(time.Now).
			Comment("开始时间").
			Annotations(entsql.WithComments(true)),

		field.Time("finished_at").
			Optional().
			Nillable().
			Comment("完成时间").
			Annotations(entsql.WithComments(true)),

		field.Int64("duration").
			Optional().
			Comment("耗时(ms)").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetCronJobLogStatuss...).
			Default(schematype.CronJobLogStatusRunning).
			Comment("状态").
			Annotations(entsql.WithComments(true)),

		field.Text("error_info").
			Optional().
			Comment("错误信息").
			Annotations(entsql.WithComments(true)),

		field.Text("output").
			Optional().
			Comment("执行输出结果").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CronJobLog.
func (CronJobLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the CronJobLog.
func (CronJobLog) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按定时任务序号排序
		index.Fields("cron_job_id").
			Annotations(entsql.WithComments(true)),

		// 按开始时间排序
		index.Fields("started_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按状态排序
		index.Fields("status").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
