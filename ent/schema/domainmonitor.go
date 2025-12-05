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
			StorageKey("id").
			Immutable().
			Comment("域名监控记录序号").
			Annotations(entsql.WithComments(true)),

		field.String("name").
			Comment("名称").
			Annotations(entsql.WithComments(true)),

		field.String("domain").
			Comment("域名").
			Annotations(entsql.WithComments(true)),

		field.Int("check_interval").
			Default(60).
			Comment("检查间隔（秒）").
			Annotations(entsql.WithComments(true)),

		field.Enum("check_protocol").
			Values(schematype.GetDomainMonitorCheckProtocol()...).
			Comment("检查类型（HTTP/HTTPS）").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_active").
			Default(false).
			Comment("是否活跃").
			Annotations(entsql.WithComments(true)),

		field.Enum("current_status").
			Values(schematype.GetDomainMonitorStatus()...).
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
	return []ent.Edge{
		// 一对多 DomainMonitor -> DomainMonitorHistory
		edge.To("history_records", DomainMonitorHistory.Type).
			Comment("关联的监控历史记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 DomainMonitor 时，其所有 DomainMonitorHistory 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the DomainMonitor.
func (DomainMonitor) Indexes() []ent.Index {
	return []ent.Index{
		// 按域名排序
		index.Fields("domain").Unique(),

		// 确保一个域名不能存在多个相同类型的记录
		index.Fields("domain", "check_protocol").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}
