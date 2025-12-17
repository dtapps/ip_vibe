package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// HttpLog holds the schema definition for the HttpLog entity.
type HttpLog struct {
	ent.Schema
}

// Annotations of the HttpLog.
func (HttpLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "http_log"},
		entsql.WithComments(true),
		schema.Comment("http 请求日志"),
	}
}

// Fields of the HttpLog.
func (HttpLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("hostname").
			Comment("请求主机名").
			Annotations(entsql.WithComments(true)),

		field.String("method").
			Comment("请求方式").
			Annotations(entsql.WithComments(true)),

		field.String("url").
			Comment("请求url").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_headers", map[string][]string{}).
			Comment("请求header").
			Annotations(entsql.WithComments(true)),

		field.JSON("request_body", json.RawMessage{}).
			Optional().
			Comment("请求参数").
			Annotations(entsql.WithComments(true)),

		field.Int("status_code").
			Comment("响应码").
			Annotations(entsql.WithComments(true)),

		field.JSON("response_headers", map[string][]string{}).
			Comment("响应header").
			Annotations(entsql.WithComments(true)),

		field.JSON("response_body", json.RawMessage{}).
			Optional().
			Comment("响应结果").
			Annotations(entsql.WithComments(true)),

		field.Int64("elapse_time").Comment("请求耗时, 单位:毫秒").
			Annotations(entsql.WithComments(true)),

		field.Bool("is_error").
			Optional().
			Default(false).
			Comment("是否为错误请求（如 5xx、4xx）").
			Annotations(entsql.WithComments(true)),

		field.Time("created_at").
			Default(time.Now).
			Comment("请求时间").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the HttpLog.
func (HttpLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the HttpLog.
func (HttpLog) Indexes() []ent.Index {
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
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the HttpLog.
func (HttpLog) Hooks() []ent.Hook {
	return []ent.Hook{}
}
