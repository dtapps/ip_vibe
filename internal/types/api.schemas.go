package types

import "encoding/json"

type SCHEMAFieldOption struct {
	Value any    `json:"value"` // 选项值
	Label string `json:"label"` // 选项标签
}

type SCHEMAField struct {
	ID          string              `json:"id"`                    // 字段ID
	Label       string              `json:"label"`                 // 字段名称
	Type        string              `json:"type"`                  // 字段类型
	ValueType   string              `json:"value_type,omitempty"`  // 字段值类型
	Placeholder string              `json:"placeholder,omitempty"` // 字段描述
	Required    bool                `json:"required"`              // 是否必填
	Pattern     string              `json:"pattern,omitempty"`     // 正则表达式
	Title       string              `json:"title,omitempty"`       // 验证提示题
	Options     []SCHEMAFieldOption `json:"options,omitempty"`     // 选项列表
}

// 获取平台配置Schema 请求结构
type GetPlatformConfigSchemaRequest struct {
	ConfigType string `uri:"config_type" json:"config_type" form:"config_type" binding:"required" jsonschema:"平台类型"` // 平台类型
}

// 获取平台配置Schema 响应结构
type GetPlatformConfigSchemaResponse struct {
	Fields []json.RawMessage `json:"fields"` // 字段列表
}

// 获取CDN记录Schema 请求结构
type GetCdnRecordSchemaRequest struct {
	GetPlatformConfigSchemaRequest
	ConfigID int64 `uri:"config_id" json:"config_id" form:"config_id" binding:"required" jsonschema:"平台配置ID"` // 平台配置ID
}

// 获取CDN记录Schema 响应结构
type GetCdnRecordSchemaResponse struct {
	Fields []json.RawMessage `json:"fields"` // 字段列表
}

// 获取DNS记录Schema 请求结构
type GetDnsRecordSchemaRequest struct {
	GetPlatformConfigSchemaRequest
	ConfigID int64 `uri:"config_id" json:"config_id" form:"config_id" binding:"required" jsonschema:"平台配置ID"` // 平台配置ID
}

// 获取DNS记录Schema 响应结构
type GetDnsRecordSchemaResponse struct {
	Fields []json.RawMessage `json:"fields"` // 字段列表
}

// 获取TUN记录Schema 请求结构
type GetTunRecordSchemaRequest struct {
	GetPlatformConfigSchemaRequest
	ConfigID int64 `uri:"config_id" json:"config_id" form:"config_id" binding:"required" jsonschema:"平台配置ID"` // 平台配置ID
}

// 获取TUN记录Schema 响应结构
type GetTunRecordSchemaResponse struct {
	Fields []json.RawMessage `json:"fields"` // 字段列表
}
