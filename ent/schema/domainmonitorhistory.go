package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// DomainMonitorHistory holds the schema definition for the DomainMonitorHistory entity.
type DomainMonitorHistory struct {
	ent.Schema
}

// Annotations of the DomainMonitorHistory.
func (DomainMonitorHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "domain_monitor_history"},
		schema.Comment("域名监控历史记录"),
		entsql.WithComments(true),
	}
}

// Fields of the DomainMonitorHistory.
func (DomainMonitorHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Immutable().
			Comment("监控历史记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("monitor_id").
			Immutable().
			Comment("关联的 DomainMonitor ID").
			Annotations(entsql.WithComments(true)),

		field.Time("check_time").
			Comment("检查发生的时间").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_success").
			Comment("本次检查是否成功").
			Annotations(entsql.WithComments(true)),

		field.Int("check_duration").
			Optional().
			Comment("检查耗时（毫秒）").
			Annotations(entsql.WithComments(true)),

		field.String("status_message").
			Optional().
			Comment("简要状态信息（如：200 OK，连接超时）").
			Annotations(entsql.WithComments(true)),

		field.Int("http_status").
			Optional().
			Comment("HTTP/HTTPS 状态码").
			Annotations(entsql.WithComments(true)),

		field.JSON("ip_resolved", []string{}).
			Optional().
			Comment("解析到的 IP 地址列表").
			Annotations(entsql.WithComments(true)),

		field.Text("error_log").
			Optional().
			Comment("详细错误信息或异常堆栈").
			Annotations(entsql.WithComments(true)),

		field.Time("ssl_expiry_date").
			Optional().
			Comment("证书过期日期").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_common_name").
			Optional().
			Comment("证书主体通用名 (Subject CN)").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_issuer_name").
			Optional().
			Comment("证书颁发者通用名 (Issuer CN)").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_signature_alg").
			Optional().
			Comment("证书签名算法").
			Annotations(entsql.WithComments(true)),

		field.Text("ssl_san_list").
			Optional().
			Comment("证书主体别名列表 (SANs，用逗号分隔)").
			Annotations(entsql.WithComments(true)),

		field.Text("ssl_certificate_chain").
			Optional().
			Comment("完整的 SSL 证书链信息 (PEM 格式文本)").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DomainMonitorHistory.
func (DomainMonitorHistory) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 DomainMonitor -> DomainMonitorHistory
		edge.From("monitor", DomainMonitor.Type).
			Ref("history_records").
			Field("monitor_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的域名监控").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the DomainMonitorHistory.
func (DomainMonitorHistory) Indexes() []ent.Index {
	return []ent.Index{
		// 快速查询某个 Monitor 的历史记录，按时间倒序
		index.Fields("monitor_id", "check_time").
			Annotations(entsql.WithComments(true)),
	}
}
