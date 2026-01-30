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

// EsaSiteRecord holds the schema definition for the EsaSiteRecord entity.
type EsaSiteRecord struct {
	ent.Schema
}

// Annotations of the EsaSiteRecord.
func (EsaSiteRecord) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "esa_site_record"},
		schema.Comment("ESA 站点记录"),
		entsql.WithComments(true),
	}
}

// Mixin of the EsaSiteRecord.
func (EsaSiteRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EsaSiteRecord.
func (EsaSiteRecord) Fields() []ent.Field {
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

		field.Enum("biz_name").
			Values(schematype.GetEsaSiteRecordBizNames...).
			Optional().
			Comment("记录加速时的业务场景").
			Annotations(entsql.WithComments(true)),

		field.JSON("data", json.RawMessage{}).
			Optional().
			Comment("记录的具体 DNS 信息").
			Annotations(entsql.WithComments(true)),

		field.Enum("proxied").
			Values(schematype.GetEsaSiteRecordProxieds...).
			Optional().
			Comment("记录是否开启代理加速").
			Annotations(entsql.WithComments(true)),

		field.String("record_id").
			Comment("记录 ID").
			Annotations(entsql.WithComments(true)),

		field.Enum("record_source_type").
			Values(schematype.GetEsaSiteRecordRecordSourceTypes...).
			Optional().
			Comment("CNAME 记录的源站类型").
			Annotations(entsql.WithComments(true)),

		field.String("record_name").
			Optional().
			Comment("记录名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("record_type").
			Values(schematype.GetEsaSiteRecordRecordTypes...).
			Optional().
			Comment("记录的 DNS 类型").
			Annotations(entsql.WithComments(true)),

		field.Int64("ttl").
			Optional().
			Comment("记录的过期时间").
			Annotations(entsql.WithComments(true)),

		field.String("record_cname").
			Optional().
			Comment("记录的 CNAME").
			Annotations(entsql.WithComments(true)),

		field.String("comment").
			Optional().
			Comment("记录的备注").
			Annotations(entsql.WithComments(true)),

		field.JSON("auth_conf", json.RawMessage{}).
			Optional().
			Comment("CNAME 记录的源站鉴权信息").
			Annotations(entsql.WithComments(true)),

		field.Enum("host_policy").
			Values(schematype.GetEsaSiteRecordHostPolicys...).
			Optional().
			Comment("回源 HOST 策略").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EsaSiteRecord.
func (EsaSiteRecord) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 EsaSite -> EsaSiteRecord
		edge.From("esa_site", EsaSite.Type).
			Ref("records").
			Field("esa_site_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的 ESA 站点").
			Annotations(entsql.WithComments(true)),
	}
}

// Indexes of the EsaSiteRecord.
func (EsaSiteRecord) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("esa_site_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EsaSiteRecord.
func (EsaSiteRecord) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EsaSiteRecordLogHook(),
	}
}
