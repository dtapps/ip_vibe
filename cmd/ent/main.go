package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {

	opts := []entc.Option{
		// entc.Dependency(
		// 	entc.DependencyType(&http.Client{}),
		// ),
		entc.TemplateFiles(
			"./ent/template/debug.tmpl",
			// "./ent/template/stringer.tmpl",
			"./ent/template/aggregate.tmpl",
			"./ent/template/order_by.tmpl",
		),
	}
	err := entc.Generate("./ent/schema", &gen.Config{
		Target: "./ent",
		Features: []gen.Feature{
			// gen.FeatureSnapshot,
			// gen.FeaturePrivacy,
			// gen.FeatureEntQL,
			gen.FeatureNamedEdges,
			gen.FeatureBidiEdgeRefs,
			// gen.FeatureSchemaConfig,
			// gen.FeatureLock,
			gen.FeatureModifier,
			gen.FeatureExecQuery,
			gen.FeatureUpsert,
		},
	}, opts...)
	if err != nil {
		log.Fatalf("代码生成错误: %v", err)
	}

	log.Println("代码生成成功")
}
