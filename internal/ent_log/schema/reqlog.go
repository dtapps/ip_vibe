package schema

import (
	"encoding/json"
	"runtime"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"go.dtapp.net/library/contrib/req_log"
)

// ReqLog holds the schema definition for the ReqLog entity.
type ReqLog struct {
	ent.Schema
}

// Annotations of the ReqLog.
func (ReqLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "req_log"},
		entsql.WithComments(true),
		schema.Comment("req 请求日志"),
	}
}

// Fields of the ReqLog.
func (ReqLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id").
			Positive(). // 正数
			Immutable().
			Comment("请求日志序号").
			Annotations(entsql.WithComments(true)), // 用 id 覆盖 框架的 id

		field.String("hostname").
			Immutable().
			Comment("主机名").
			Annotations(entsql.WithComments(true)),

		field.String("method").
			Immutable().
			Comment("请求方法").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Immutable().
			Comment("请求 URL").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_headers", map[string][]string{}).
			Optional().
			Immutable().
			Comment("请求头").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_body", json.RawMessage{}).
			Optional().
			Immutable().
			Comment("请求体").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_query_params", map[string][]string{}).
			Optional().
			Immutable().
			Comment("请求参数").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_form_data", map[string][]string{}).
			Optional().
			Immutable().
			Comment("请求表单数据").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_path_params", map[string]string{}).
			Optional().
			Immutable().
			Comment("请求路径参数").
			Annotations(entsql.WithComments(true)),

		field.Int("status_code").
			Comment("状态码").
			Optional().
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.JSON("response_headers", map[string][]string{}).
			Optional().
			Immutable().
			Comment("响应头").
			Annotations(entsql.WithComments(true)),

		field.JSON("response_body", json.RawMessage{}).
			Optional().
			Immutable().
			Comment("响应体").
			Annotations(entsql.WithComments(true)),

		field.Int64("elapse_time").
			Optional().
			Immutable().
			Comment("耗时, 单位:毫秒").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_error").
			Default(false).
			Optional().
			Immutable().
			Comment("是否错误（如 5xx、4xx）").
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("请求时间").
			Annotations(entsql.WithComments(true)),

		field.String("go_version").
			Default(runtime.Version()).
			Optional().
			Immutable().
			Comment("Go 版本").
			Annotations(entsql.WithComments(true)),

		field.String("plugin_version").
			Default(req_log.Version).
			Optional().
			Immutable().
			Comment("插件版本").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the ReqLog.
func (ReqLog) Edges() []ent.Edge {
	return []ent.Edge{}
}

// Indexes of the ReqLog.
func (ReqLog) Indexes() []ent.Index {
	return []ent.Index{
		// 按请求主机名排序
		index.Fields("hostname").
			Annotations(entsql.WithComments(true)),

		// 按请求方式排序
		index.Fields("method").
			Annotations(entsql.WithComments(true)),

		// 按请求url排序
		index.Fields("url").
			Annotations(entsql.WithComments(true)),

		// 按响应状态码排序
		index.Fields("status_code").
			Annotations(entsql.WithComments(true)),

		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),
	}
}

// Hooks of the ReqLog.
func (ReqLog) Hooks() []ent.Hook {
	return []ent.Hook{}
}
