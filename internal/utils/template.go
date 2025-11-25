package utils

import (
	"bytes"
	"html/template"
	"log"
)

// RenderTemplate 使用 Go Template 渲染标题和内容
func RenderTemplate(tmplString string, data any) (string, error) {
	// 创建一个新的模板实例
	tmpl := template.New("webhook_template")

	// 将默认的 {{ 和 }} 改为 [[ 和 ]]
	tmpl.Delims("[[", "]]")

	// 解析模板字符串
	tmpl, err := tmpl.Parse(tmplString)
	if err != nil {
		log.Printf("解析模板失败: %v", err)
		return "", err
	}

	var buf bytes.Buffer

	// 执行渲染
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("渲染模板失败: %v", err)
		return "", err
	}

	return buf.String(), nil
}
