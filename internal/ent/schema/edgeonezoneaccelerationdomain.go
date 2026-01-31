package schema

import (
	"encoding/json"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent/schema/schemamixin"
	"vibe_ip/internal/ent/schema/schematype"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// EdgeOneZoneAccelerationDomain holds the schema definition for the EdgeOneZoneAccelerationDomain entity.
type EdgeOneZoneAccelerationDomain struct {
	ent.Schema
}

// Annotations of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "edgeone_zone_accelerationdomain"},
		schema.Comment("EdgeOne 站点加速域名"),
		entsql.WithComments(true),
	}
}

// Mixin of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("EdgeOne 站点加速域名序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("edgeone_zone_id").
			Positive(). // 正数
			Immutable().
			Comment("关联 EdgeOne 站点序号").
			Annotations(entsql.WithComments(true)),

		field.String("domain_name").
			Immutable().
			Comment("加速域名名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("domain_status").
			Values(schematype.GetEdgeOneZoneAccelerationDomainDomainStatuss...).
			Optional().
			Comment("加速域名状态").
			Annotations(entsql.WithComments(true)),

		field.String("cname").
			Optional().
			Comment("CNAME 地址").
			Annotations(entsql.WithComments(true)),

		field.Enum("ipv6_status").
			Values(schematype.GetEdgeOneZoneAccelerationDomainIPv6Statuss...).
			Optional().
			Comment("IPv6 状态").
			Annotations(entsql.WithComments(true)),

		field.Enum("identification_status").
			Values(schematype.GetEdgeOneZoneAccelerationDomainIdentificationStatuss...).
			Optional().
			Comment("加速域名归属权验证状态").
			Annotations(entsql.WithComments(true)),

		field.JSON("ownership_verification", json.RawMessage{}).
			Optional().
			Comment("加速域名需进行归属权验证才能继续提供服务时，该对象会携带对应验证方式所需要的信息").
			Annotations(entsql.WithComments(true)),

		field.JSON("origin_detail", json.RawMessage{}).
			Optional().
			Comment("源站信息").
			Annotations(entsql.WithComments(true)),

		field.Enum("origin_protocol").
			Values(schematype.GetEdgeOneZoneAccelerationDomainOriginProtocols...).
			Optional().
			Comment("回源协议").
			Annotations(entsql.WithComments(true)),

		field.Uint64("http_origin_port").
			Optional().
			Comment("HTTP 回源端口").
			Annotations(entsql.WithComments(true)),

		field.Uint64("https_origin_port").
			Optional().
			Comment("HTTPS 回源端口").
			Annotations(entsql.WithComments(true)),

		field.JSON("certificate", json.RawMessage{}).
			Optional().
			Comment("加速域名证书信息").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EdgeOneZone -> EdgeOneZoneAccelerationDomain
		edge.From("edgeone_zone", EdgeOneZone.Type).
			Ref("accelerationdomains").
			Field("edgeone_zone_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 EdgeOne 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("edgeone_zone_id").
			Annotations(entsql.WithComments(true)),

		// 按加速域名名称排序
		index.Fields("domain_name").
			Annotations(entsql.WithComments(true)),

		// 按站点ID+加速域名名称排序
		index.Fields("edgeone_zone_id", "domain_name").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EdgeOneZoneAccelerationDomain.
func (EdgeOneZoneAccelerationDomain) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EdgeOneZoneAccelerationDomainLogHook(),
	}
}
