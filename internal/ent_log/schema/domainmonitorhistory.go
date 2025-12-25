package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
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

// Mixin of the DomainMonitorHistory.
func (DomainMonitorHistory) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the DomainMonitorHistory.
func (DomainMonitorHistory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("域名监控历史记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("monitor_id").
			Immutable().
			Comment("关联 DomainMonitor 序号").
			Annotations(entsql.WithComments(true)),

		field.Time("check_time").
			Immutable().
			Comment("检查发生的时间").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_success").
			Immutable().
			Comment("本次检查是否成功").
			Annotations(entsql.WithComments(true)),

		field.Int("check_duration").
			Optional().
			Immutable().
			Comment("检查耗时（毫秒）").
			Annotations(entsql.WithComments(true)),

		field.String("status_message").
			Optional().
			Immutable().
			Comment("简要状态信息（如：200 OK，连接超时）").
			Annotations(entsql.WithComments(true)),

		field.Int("http_status").
			Optional().
			Immutable().
			Comment("HTTP/HTTPS 状态码").
			Annotations(entsql.WithComments(true)),

		field.JSON("ip_resolved", []string{}).
			Optional().
			Immutable().
			Comment("解析到的 IP 地址列表").
			Annotations(entsql.WithComments(true)),

		field.Text("error_log").
			Optional().
			Immutable().
			Comment("详细错误信息或异常堆栈").
			Annotations(entsql.WithComments(true)),

		field.Time("ssl_not_date").
			Optional().
			Immutable().
			Comment("证书生效日期（NotBefore）").
			Annotations(entsql.WithComments(true)),

		field.Time("ssl_expiry_date").
			Optional().
			Immutable().
			Comment("证书过期日期（NotAfter）").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_common_name").
			Optional().
			Immutable().
			Comment("证书主体通用名 (Subject CommonName)").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_issuer_name").
			Optional().
			Immutable().
			Comment("证书颁发者通用名 (Issuer CommonName)").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_signature_alg").
			Optional().
			Immutable().
			Comment("证书签名算法 (SignatureAlgorithm)").
			Annotations(entsql.WithComments(true)),

		field.Text("ssl_san_list").
			Optional().
			Immutable().
			Comment("证书主体别名列表 (DNSNames，用逗号分隔)").
			Annotations(entsql.WithComments(true)),

		field.Text("ssl_certificate_chain").
			Optional().
			Immutable().
			Comment("完整的 SSL 证书链信息 (PEM格式)").
			Annotations(entsql.WithComments(true)),

		field.String("ssl_serial_number").
			Optional().
			Immutable().
			Comment("序列号 (SerialNumber)").
			Annotations(entsql.WithComments(true)),

		field.Int("ssl_version").
			Optional().
			Immutable().
			Comment("版本 (Version)").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the DomainMonitorHistory.
func (DomainMonitorHistory) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the DomainMonitorHistory.
func (DomainMonitorHistory) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按关联 DomainMonitor 序号分组
		index.Fields("monitor_id").
			Annotations(entsql.WithComments(true)),

		// 按检查发生的时间排序
		index.Fields("check_time").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 快速查询某个 Monitor 的历史记录，按时间倒序
		index.Fields("monitor_id", "check_time").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}
