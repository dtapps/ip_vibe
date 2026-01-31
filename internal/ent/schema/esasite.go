package schema

import (
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

// EsaSite holds the schema definition for the EsaSite entity.
type EsaSite struct {
	ent.Schema
}

// Annotations of the EsaSite.
func (EsaSite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "esa_site"},
		schema.Comment("ESA 站点"),
		entsql.WithComments(true),
	}
}

// Mixin of the EsaSite.
func (EsaSite) Mixin() []ent.Mixin {
	return []ent.Mixin{
		schemamixin.AtMixin{},
		schemamixin.SyncAtMixin{},
	}
}

// Fields of the EsaSite.
func (EsaSite) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			StorageKey("id"). // 用 id 覆盖 框架的 id
			Positive().       // 正数
			Immutable().
			Comment("ESA 站点序号").
			Annotations(entsql.WithComments(true)),

		field.Int64("config_id").
			Positive(). // 正数
			Immutable().
			Comment("平台配置序号").
			Annotations(entsql.WithComments(true)),

		field.Enum("access_type").
			Values(schematype.GetEsaSiteAccessTypes...).
			Optional().
			Comment("站点接入类型").
			Annotations(entsql.WithComments(true)),

		field.String("cname_zone").
			Optional().
			Comment("站点的 CNAME 后缀").
			Annotations(entsql.WithComments(true)),

		field.Enum("coverage").
			Values(schematype.GetEsaSiteCoverages...).
			Optional().
			Comment("站点加速区域").
			Annotations(entsql.WithComments(true)),

		field.String("instance_id").
			Optional().
			Comment("站点绑定的套餐实例 ID").
			Annotations(entsql.WithComments(true)),

		field.String("name_server_list").
			Optional().
			Comment("站点分配的 NS 列表。使用半角逗号（,）分隔").
			Annotations(entsql.WithComments(true)),

		field.String("plan_name").
			Optional().
			Comment("套餐名称").
			Annotations(entsql.WithComments(true)),

		field.String("plan_spec_name").
			Optional().
			Comment("站点套餐的规格名称").
			Annotations(entsql.WithComments(true)),

		field.Int64("site_id").
			Immutable().
			Comment("站点 ID").
			Annotations(entsql.WithComments(true)),

		field.String("site_name").
			Optional().
			Comment("站点名称").
			Annotations(entsql.WithComments(true)),

		field.Enum("status").
			Values(schematype.GetEsaSiteStatuss...).
			Optional().
			Comment("站点状态").
			Annotations(entsql.WithComments(true)),

		field.String("offline_reason").
			Optional().
			Comment("站点停用原因").
			Annotations(entsql.WithComments(true)),
	}
}

// Edges of the EsaSite.
func (EsaSite) Edges() []ent.Edge {
	return []ent.Edge{
		// 一对多 PlatformConfig -> EsaSite
		edge.From("config", PlatformConfig.Type).
			Ref("esa_zones").
			Field("config_id").
			Unique().
			Required().
			Immutable().
			Comment("关联的平台配置").
			Annotations(entsql.WithComments(true)),

		// 一对多 EsaSite -> EsaSiteRecord
		edge.To("records", EsaSiteRecord.Type).
			Comment("关联的ESA站点记录").
			Annotations(entsql.Annotation{
				// 级联删除：删除 EsaSite 时，其所有 EsaSiteRecord 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),

		// 一对多 EsaSite -> EsaSiteOriginPool
		edge.To("originpools", EsaSiteOriginPool.Type).
			Comment("关联的ESA站点源地址池").
			Annotations(entsql.Annotation{
				// 级联删除：删除 EsaSite 时，其所有 EsaSiteOriginPool 记录也应删除
				OnDelete: entsql.Cascade,
			}, entsql.WithComments(true)),
	}
}

// Indexes of the EsaSite.
func (EsaSite) Indexes() []ent.Index {
	return []ent.Index{
		// 按创建时间排序
		index.Fields("created_at").
			Annotations(entsql.Desc(), entsql.WithComments(true)),

		// 按平台配置序号排序
		index.Fields("config_id").
			Annotations(entsql.WithComments(true)),

		// 按站点ID排序
		index.Fields("site_id").
			Annotations(entsql.WithComments(true)),

		// 按平台配置序号+站点ID排序
		index.Fields("config_id", "site_id").
			Annotations(entsql.WithComments(true)),
	}
}

// Hooks of the EsaSite.
func (EsaSite) Hooks() []ent.Hook {
	return []ent.Hook{
		schemahook.EsaSiteLogHook(),
	}
}
