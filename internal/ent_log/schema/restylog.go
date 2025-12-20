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
	"go.dtapp.net/library/contrib/resty_log"
	"resty.dev/v3"
)

// RestyLog holds the schema definition for the RestyLog entity.
type RestyLog struct {
	ent.Schema
}

// Annotations of the RestyLog.
func (RestyLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "resty_log"},
		entsql.WithComments(true),
		schema.Comment("resty 请求日志"),
	}
}

// Fields of the RestyLog.
func (RestyLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("hostname").
			Comment("请求主机名").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.String("method").
			Comment("请求方式").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Comment("请求url").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.JSON("request_headers", map[string][]string{}).
			Comment("请求header").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.JSON("request_body", json.RawMessage{}).
			Optional().
			Comment("请求参数").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.Int("status_code").
			Comment("响应码").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.JSON("response_headers", map[string][]string{}).
			Comment("响应header").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.JSON("response_body", json.RawMessage{}).
			Optional().
			Comment("响应结果").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.Int64("elapse_time").
			Comment("请求耗时, 单位:毫秒").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.Bool("is_error").
			Optional().
			Default(false).
			Comment("是否为错误请求（如 5xx、4xx）").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Comment("请求时间").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.String("go_version").
			Default(runtime.Version()).
			Optional().
			Comment("go 版本").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.String("plugin_version").
			Default(resty_log.Version).
			Optional().
			Comment("plugin 版本").
			Immutable().
			Annotations(entsql.WithComments(true)),

		field.String("resty_version").
			Default(resty.Version).
			Optional().
			Comment("resty 版本").
			Immutable().
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the RestyLog.
func (RestyLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the RestyLog.
func (RestyLog) Indexes() []ent.Index {
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

// Hooks of the RestyLog.
func (RestyLog) Hooks() []ent.Hook {
	return []ent.Hook{}
}
