package utils

import (
	"bytes"
	"html/template"
	"log"
)

type TemplateData struct {
	EventTitle     string // 事件标题
	EventType      string // 事件类型
	EventTime      string // 事件时间
	EventBody      string // 事件内容
	DeviceCode     string // 设备编码 Device->Code
	DeviceName     string // 设备名称 Device->Name
	IPv4Addr       string // IP地址V4 Device->IPv4
	IPv6Addr       string // IP地址V6 Device->IPv6
	RecordDomain   string // 记录域名 CDNRecord/DNSRecord->Domain
	RecordType     string // 记录类型 CDNRecord/DNSRecord->Type
	RecordOldValue string // 旧记录值 CDNRecord/DNSRecord->Record
	RecordNewValue string // 新记录值 CDNRecord/DNSRecord->Record
	PlatformName   string // 平台名称 PlatformConfig->Name
	PlatformType   string // 平台类型 PlatformConfig->Platform
}

// RenderTemplate 使用 Go Template 渲染标题和内容
func RenderTemplate(tmplString string, data TemplateData) (string, error) {
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
