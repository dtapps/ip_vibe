package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// CloudflareZoneAnalysisData holds the schema definition for the CloudflareZoneAnalysisData entity.
type CloudflareZoneAnalysisData struct {
	ent.Schema
}

// Annotations of the CloudflareZoneAnalysisData.
func (CloudflareZoneAnalysisData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "cloudflare_zone_analysis_data"},
		schema.Comment("Cloudflare 站点分析数据"),
		entsql.WithComments(true),
	}
}

// Mixin of the CloudflareZoneAnalysisData.
func (CloudflareZoneAnalysisData) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the CloudflareZoneAnalysisData.
func (CloudflareZoneAnalysisData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点分析数据序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("cloudflare_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 Cloudflare 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Time("date").
			Immutable().
			Comment("分析日期").
			Annotations(entsql.WithComments(true)),

		field.Int64("query_count").
			Optional().
			Comment("查询计数").
			Annotations(entsql.WithComments(true)),

		field.Int64("uncached_count").
			Optional().
			Comment("未缓存查询计数").
			Annotations(entsql.WithComments(true)),

		field.Int64("stale_count").
			Optional().
			Comment("未缓存查询计数").
			Annotations(entsql.WithComments(true)),

		field.Float("response_time_avg").
			Optional().
			Comment("响应耗时平均值，单位：ms").
			Annotations(entsql.WithComments(true)),

		field.Float("response_time_median").
			Optional().
			Comment("响应耗时中位数，单位：ms").
			Annotations(entsql.WithComments(true)),

		field.Float("response_time_90th").
			Optional().
			Comment("响应耗时90分位数，单位：ms").
			Annotations(entsql.WithComments(true)),

		field.Float("response_time_99th").
			Optional().
			Comment("响应耗时99分位数，单位：ms").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the CloudflareZoneAnalysisData.
func (CloudflareZoneAnalysisData) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("cloudflare_zone_id").
			Annotations(entsql.WithComments(true)),

		// 按分析日期排序
		index.Fields("date").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID+分析日期排序
		index.Fields("cloudflare_zone_id", "date").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the CloudflareZoneAnalysisData.
func (CloudflareZoneAnalysisData) Edges() []ent.Edge {
	return nil
}
