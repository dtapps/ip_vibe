package main

import (
	"fmt"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {

	target := "./ent"
	target = "./internal/ent"
	// target = "./internal/ent_log"
	schemaPath := target + "/schema"

	fmt.Println("target", target)
	fmt.Println("schemaPath", schemaPath)

	opts := []entc.Option{
		// entc.Dependency(
		// 	entc.DependencyType(&http.Client{}),
		// ),
		entc.TemplateFiles(
			"./internal/ent/template/debug.tmpl",
			// "./internal/ent/template/stringer.tmpl",
			"./internal/ent/template/aggregate.tmpl",
			"./internal/ent/template/order_by.tmpl",
			"./internal/ent/template/pagination.tmpl",
			"./internal/ent/template/to_map.tmpl",
		),
	}
	err := entc.Generate(schemaPath, &gen.Config{
		Target: target,
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
		fmt.Printf("代码生成错误: %v", err)
		return
	}

	fmt.Println("代码生成成功")
}
