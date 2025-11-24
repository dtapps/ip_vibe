package types

import "encoding/json"

// 获取全部配置 响应结构
type GetConfigsResponse struct {
	ID         int64           `json:"id,omitempty"`          // 配置序号
	Capability string          `json:"capability"`            // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name       string          `json:"name"`                  // 配置名称
	Platform   string          `json:"platform"`              // 配置平台
	ConfigJSON json.RawMessage `json:"config_json,omitempty"` // 配置平台密钥等敏感信息（JSON）
}

// 创建配置 请求结构
type CreateConfigRequest struct {
	Capability string `json:"capability"`         // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name       string `json:"name"`               // 配置名称
	Platform   string `json:"platform,omitempty"` // 配置平台
	ConfigJSON string `json:"config_json"`        // 配置平台密钥等敏感信息（JSON）
}

// 更新配置 请求结构
type UpdateConfigRequest struct {
	Name       string `json:"name"`        // 配置名称
	ConfigJSON string `json:"config_json"` // 配置平台密钥等敏感信息（JSON）
}
