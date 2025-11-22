package types

import "encoding/json"

// 配置 响应结构
type GetConfigsResponse struct {
	ID         int64           `json:"id,omitempty"`          // 配置序号
	Name       string          `json:"name"`                  // 配置名称
	Platform   string          `json:"platform"`              // 配置平台
	ConfigJSON json.RawMessage `json:"config_json,omitempty"` // 配置平台密钥等敏感信息（JSON）
}

// 配置 请求结构
type CreateConfigRequest struct {
	ID         int64  `json:"id,omitempty"`       // 配置序号
	Name       string `json:"name"`               // 配置名称
	Platform   string `json:"platform,omitempty"` // 配置平台
	ConfigJSON string `json:"config_json"`        // 配置平台密钥等敏感信息（JSON）
}
