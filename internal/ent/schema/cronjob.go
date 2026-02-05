package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemadefault"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CronJob holds the schema definition for the CronJob entity.
type CronJob struct {
	ent.Schema
}

// Annotations of the CronJob.
func (CronJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cron_job"},
		schema.Comment("定时任务"),
		entsql.WithComments(true),
	}
}

// Mixin of the CronJob.
func (CronJob) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the CronJob.
func (CronJob) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("任务序号").
			Annotations(entsql.WithComments(true)),

		field.UUID("entry_id", uuid.UUID{}).
			Default(schemadefault.DefaultUUID).
			Unique().
			Immutable().
			Comment("任务编号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			NotEmpty().
			Comment("任务名称").
			Annotations(entsql.WithComments(true)),

		field.String("handler_name").
			NotEmpty().
			Comment("任务处理函数").
			Annotations(entsql.WithComments(true)),

		field.JSON("handler_args", json.RawMessage{}).
			Optional().
			Comment("任务处理函数参数").
			Annotations(entsql.WithComments(true)),

		field.Enum("task_type").
			Values(schematype.GetCronJobTaskTypes...).
			Default(schematype.CronJobTaskTypeRecurring).
			Comment("任务类型: one_off 一次性, recurring 周期性").
			Annotations(entsql.WithComments(true)),

		field.Enum("job_type").
			Values(schematype.GetCronJobJobTypes...).
			Default(schematype.CronJobJobTypeCron).
			Comment("任务类型: cron cron 表达式, duration 固定间隔, random 随机间隔").
			Annotations(entsql.WithComments(true)),

		field.Enum("job_category").
			Values(schematype.GetCronJobJobCategorys...).
			Comment("任务分类: system 系统内置, custom 用户自定义").
			Annotations(entsql.WithComments(true)),

		field.Enum("business_type").
			Values(schematype.GetCronJobBusinessTypes...).
			Optional().
			Comment("业务类型: cdn, dns, ssl, system").
			Annotations(entsql.WithComments(true)),

		field.Int64("business_id").
			Optional().
			Comment("业务序号"),

		field.String("cron_expr").
			Optional().
			Comment("cron 表达式").
			Annotations(entsql.WithComments(true)),

		field.Bool("with_seconds").
			Default(false).
			Comment("是否包含秒").
			Annotations(entsql.WithComments(true)),

		field.Int("interval_value").
			Optional().
			Comment("固定间隔值 (配合 interval_unit 使用)").
			Annotations(entsql.WithComments(true)),

		field.Int("min_interval").
			Optional().
			Min(1).
			Comment("随机最小值 (配合 interval_unit 使用)").
			Annotations(entsql.WithComments(true)),

		field.Int("max_interval").
			Optional().
			Comment("随机最大值 (配合 interval_unit 使用)").
			Annotations(entsql.WithComments(true)),

		field.Enum("interval_unit").
			Values(schematype.GetCronJobIntervalUnits...).
			Default(schematype.CronJobIntervalUnitMinutes).
			Comment("间隔时间单位 (用于 duration 和 random 模式)").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetCronJobStatuss...).
			Default(schematype.CronJobStatusStopped).
			Comment("状态: running 运行中, stopped 已停止").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_singleton").
			Default(true).
			Comment("是否单例").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CronJob.
func (CronJob) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the CronJob.
func (CronJob) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按任务类型排序
		index.Fields("job_type").
			Annotations(entsql.WithComments(true)),

		// 按任务分类排序
		index.Fields("job_category").
			Annotations(entsql.WithComments(true)),

		// 按业务类型排序
		index.Fields("business_type").
			Annotations(entsql.WithComments(true)),

		// 按状态排序
		index.Fields("status").
			Annotations(entsql.WithComments(true)),

		// 按业务类型+业务序号排序
		index.Fields("business_type", "business_id").
			Annotations(entsql.WithComments(true)),

		// 按任务分类+业务类型+业务序号排序
		index.Fields("job_category", "business_type", "business_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the CronJob.
func (CronJob) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.CronJobLogHook(),
	}
}
