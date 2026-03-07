package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EsaSiteAnalysisData holds the schema definition for the EsaSiteAnalysisData entity.
type EsaSiteAnalysisData struct {
	ent.Schema
}

// Annotations of the EsaSiteAnalysisData.
func (EsaSiteAnalysisData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "esa_site_analysis_data"},
		schema.Comment("ESA 站点分析数据"),
		entsql.WithComments(true),
	}
}

// Mixin of the EsaSiteAnalysisData.
func (EsaSiteAnalysisData) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the EsaSiteAnalysisData.
func (EsaSiteAnalysisData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("ESA 站点记录序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("esa_site_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 ESA 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Time("date").
			Immutable().
			Comment("分析日期").
			Annotations(entsql.WithComments(true)),

		field.Int64("traffic").
			Optional().
			Comment("ESA节点响应返回给客户端的大小，单位：Byte").
			Annotations(entsql.WithComments(true)),

		field.Int64("requests").
			Optional().
			Comment("请求数").
			Annotations(entsql.WithComments(true)),

		field.Int64("request_traffic").
			Optional().
			Comment("客户端请求的大小，单位：Byte").
			Annotations(entsql.WithComments(true)),

		field.Int64("page_view").
			Optional().
			Comment("页面浏览量").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EsaSiteAnalysisData.
func (EsaSiteAnalysisData) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("esa_site_id").
			Annotations(entsql.WithComments(true)),

		// 按分析日期排序
		index.Fields("date").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID+分析日期排序
		index.Fields("esa_site_id", "date").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EsaSiteAnalysisData.
func (EsaSiteAnalysisData) Edges() []ent.Edge {
	return nil
}
