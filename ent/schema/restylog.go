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
		field.String("hostname").Comment("请求主机名").Annotations(entsql.WithComments(true)),
		field.String("method").Comment("请求方式").Annotations(entsql.WithComments(true)),
		field.String("url").Comment("请求url").Annotations(entsql.WithComments(true)),
		field.JSON("request_headers", map[string][]string{}).Comment("请求header").Annotations(entsql.WithComments(true)),
		field.JSON("request_body", json.RawMessage{}).Optional().Comment("请求参数").Annotations(entsql.WithComments(true)),
		field.Int("status_code").Comment("响应码").Annotations(entsql.WithComments(true)),
		field.JSON("response_headers", map[string][]string{}).Comment("响应header").Annotations(entsql.WithComments(true)),
		field.JSON("response_body", json.RawMessage{}).Optional().Comment("响应结果").Annotations(entsql.WithComments(true)),
		field.Int64("elapse_time").Comment("请求耗时, 单位:毫秒").Annotations(entsql.WithComments(true)),
		field.Bool("is_error").Optional().Default(false).Comment("是否为错误请求（如 5xx、4xx）").Annotations(entsql.WithComments(true)),
		field.Time("created_at").Default(time.Now).Comment("请求时间").Annotations(entsql.WithComments(true)),
	}
}

// Edges of the RestyLog.
func (RestyLog) Edges() []ent.Edge {
	return nil
}

// Indexes of the RestyLog.
func (RestyLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("hostname"),
		index.Fields("method"),
		index.Fields("url"),
		index.Fields("status_code"),
	}
}

// Hooks of the RestyLog.
func (RestyLog) Hooks() []ent.Hook {
	return []ent.Hook{}
}
