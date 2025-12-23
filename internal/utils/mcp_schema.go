package utils

import (
	"reflect"
	"strings"
)

/**
InputSchema: `{
    "type": "object",
    "properties": {
        "id": {
            "type": "integer",
            "description": "CDN记录的唯一标识符"
        }
    },
    "required": ["id"]
}`
**/

// AutoSchema 自动将结构体转换为 MCP InputSchema（支持嵌套结构体展开）
func AutoSchema(obj any) map[string]any {
	return buildSchemaFromType(reflect.TypeOf(obj))
}

// 递归构建 Schema，支持嵌套结构体展开
func buildSchemaFromType(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": make(map[string]any),
		"required":   []string{},
	}

	// 收集所有字段（包括嵌套结构体的字段）
	allFields := collectAllFields(t)

	for _, fieldInfo := range allFields {
		processFieldWithPath(fieldInfo, schema)
	}

	return schema
}

// 字段信息结构
type fieldInfo struct {
	field    reflect.StructField
	name     string
	jsonName string
	path     string
}

// 收集所有字段，包括嵌套结构体的字段
func collectAllFields(t reflect.Type) []fieldInfo {
	var fields []fieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		jsonFieldName := extractJSONFieldName(jsonTag)
		if jsonFieldName == "" {
			jsonFieldName = field.Name
		}

		// 如果是嵌套结构体且不是匿名嵌入，正常处理
		if field.Anonymous {
			// 处理匿名嵌入（如 UpdateCdnRecordRequest 嵌入 GetCdnRecordRequest）
			nestedFields := collectAllFields(field.Type)
			for _, nestedField := range nestedFields {
				// 为嵌套字段添加前缀路径
				nestedField.path = field.Name + "." + nestedField.path
				fields = append(fields, nestedField)
			}
		} else {
			// 普通字段
			fields = append(fields, fieldInfo{
				field:    field,
				name:     field.Name,
				jsonName: jsonFieldName,
				path:     jsonFieldName,
			})
		}
	}

	return fields
}

// 处理字段（支持路径展开）
func processFieldWithPath(fieldInfo fieldInfo, schema map[string]any) {
	field := fieldInfo.field
	jsonFieldName := fieldInfo.jsonName

	// 获取字段类型（处理指针）
	fieldType := field.Type
	if fieldType.Kind() == reflect.Ptr {
		fieldType = fieldType.Elem()
	}

	// 创建属性
	property := map[string]any{
		"type":        getSchemaType(fieldType),
		"description": field.Tag.Get("jsonschema"),
	}

	// 处理特殊类型
	if fieldType.String() == "time.Time" {
		property["type"] = "string"
		property["description"] = "RFC3339 timestamp"
	}

	// 处理 any 类型
	if fieldType.Kind() == reflect.Interface && fieldType.Name() == "" {
		property["type"] = "string" // any 类型转为 string
	}

	// 添加到 properties
	props := schema["properties"].(map[string]any)
	props[jsonFieldName] = property

	// 判断是否为必需字段
	if isRequiredField(field.Tag.Get("json"), field.Tag.Get("binding"), field.Type) {
		required := schema["required"].([]string)
		schema["required"] = append(required, jsonFieldName)
	}
}

// 提取 JSON 字段名
func extractJSONFieldName(jsonTag string) string {
	if jsonTag == "" {
		return ""
	}

	parts := strings.Split(jsonTag, ",")
	if len(parts) > 0 && parts[0] != "" && parts[0] != "-" {
		return parts[0]
	}
	return ""
}

// 判断是否为必需字段
func isRequiredField(jsonTag, bindingTag string, fieldType reflect.Type) bool {
	// 如果有 binding:"required" 标签 → 必需
	if strings.Contains(bindingTag, "required") {
		return true
	}

	// 如果 json tag 没有 omitempty 且不是指针类型 → 必需
	if !strings.Contains(jsonTag, "omitempty") && fieldType.Kind() != reflect.Ptr {
		return true
	}

	return false
}

// 获取 Schema 类型
func getSchemaType(fieldType reflect.Type) string {
	switch fieldType.Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.String:
		return "string"
	case reflect.Struct:
		// 对于结构体，返回 object，但描述其性质
		if fieldType.String() == "time.Time" {
			return "string"
		}
		return "object"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Interface:
		// any 类型
		return "string"
	default:
		return "string"
	}
}
