package schema

import (
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DomainMonitor holds the schema definition for the DomainMonitor entity.
type DomainMonitor struct {
	ent.Schema
}

// Annotations of the DomainMonitor.
func (DomainMonitor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "domain_monitor"},
		schema.Comment("域名监控"),
		entsql.WithComments(true),
	}
}

// Mixin of the DomainMonitor.
func (DomainMonitor) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
	}
}

// Fields of the DomainMonitor.
func (DomainMonitor) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("域名监控记录序号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("域名监控名称").
			Annotations(entsql.WithComments(true)),

		field.String("description").
			Optional().
			Comment("域名监控描述").
			Annotations(entsql.WithComments(true)),

		field.String("domain").
			Immutable().
			Comment("域名").
			Annotations(entsql.WithComments(true)),

		field.Strings("request_headers").
			Optional().
			Comment("请求头（多个之间用换行符分隔，格式：key=value）").
			Annotations(entsql.WithComments(true)),

		field.Int("check_interval").
			Default(60).
			Comment("检查间隔（秒）").
			Annotations(entsql.WithComments(true)),

		field.Enum("check_protocol").
			Values(schematype.GetDomainMonitorCheckProtocols...).
			Immutable().
			Comment("检查类型（HTTP/HTTPS）").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Enum("current_status").
			Values(schematype.GetDomainMonitorStatuss...).
			Default(schematype.DomainMonitorStatusUnknown).
			Comment("当前运行状态").
			Annotations(entsql.WithComments(true)),

		field.Time("last_success_at").
			Optional().
			Comment("最后一次检查成功时间").
			Annotations(entsql.WithComments(true)),

		field.Time("last_failure_at").
			Optional().
			Comment("最后一次检查失败时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DomainMonitor.
func (DomainMonitor) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the DomainMonitor.
func (DomainMonitor) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按域名排序
		index.Fields("domain").
			Annotations(entsql.WithComments(true)),

		// 确保一个域名不能存在多个相同类型的记录
		index.Fields("domain", "check_protocol").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the DomainMonitor.
func (DomainMonitor) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.DomainMonitorLogHook(),
	}
}
