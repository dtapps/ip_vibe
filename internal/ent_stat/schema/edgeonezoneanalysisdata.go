package schema

import (
	"vibe_ip/internal/ent/schema/schemamixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EdgeOneZoneAnalysisData holds the schema definition for the EdgeOneZoneAnalysisData entity.
type EdgeOneZoneAnalysisData struct {
	ent.Schema
}

// Annotations of the EdgeOneZoneAnalysisData.
func (EdgeOneZoneAnalysisData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edge_one_zone_analysis_data"},
		schema.Comment("EdgeOne 站点分析数据"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneZoneAnalysisData.
func (EdgeOneZoneAnalysisData) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtLogMixin{},
	}
}

// Fields of the EdgeOneZoneAnalysisData.
func (EdgeOneZoneAnalysisData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点分析数据序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("edgeone_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 EdgeOne 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Time("date").
			Immutable().
			Comment("分析日期").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_out_flux").
			Optional().
			Comment("L7 EdgeOne 响应流量，单位：Byte").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_in_flux").
			Optional().
			Comment("L7 客户端请求流量，单位：Byte").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_flux").
			Optional().
			Comment("L7 访问总流量（EdgeOne 响应+客户端请求），单位：Byte").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_out_bandwidth").
			Optional().
			Comment("L7 EdgeOne 响应带宽，单位：bps").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_in_bandwidth").
			Optional().
			Comment("L7 客户端请求带宽，单位：bps").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_bandwidth").
			Optional().
			Comment("L7 访问总带宽（EdgeOne 响应+客户端请求），单位：bps").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_request").
			Optional().
			Comment("L7 访问请求数，单位：次").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_avg_response_time").
			Optional().
			Comment("L7 访问平均响应耗时，单位：ms").
			Annotations(entsql.WithComments(true)),

		field.Int64("l7_flow_avg_first_byte_response_time").
			Optional().
			Comment("L7 访问平均首字节响应耗时，单位：ms").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneZoneAnalysisData.
func (EdgeOneZoneAnalysisData) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("edgeone_zone_id").
			Annotations(entsql.WithComments(true)),

		// 按分析日期排序
		index.Fields("date").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID+分析日期排序
		index.Fields("edgeone_zone_id", "date").
			Unique().
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EdgeOneZoneAnalysisData.
func (EdgeOneZoneAnalysisData) Edges() []ent.Edge {
	return nil
}
