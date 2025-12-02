package types

import "encoding/json"

// 获取全部配置 响应结构
type GetConfigsResponse struct {
	ID          int64           `json:"id,omitempty"`          // 配置序号
	Capability  string          `json:"capability"`            // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string          `json:"name"`                  // 配置名称
	Description string          `json:"description,omitempty"` // 配置描述
	Platform    string          `json:"platform"`              // 配置平台
	ConfigJSON  json.RawMessage `json:"config_json,omitempty"` // 配置平台密钥等敏感信息（JSON）
}

// 获取配置 请求结构
type GetConfigRequest struct {
	ID int64 `uri:"id" binding:"required"` // 配置序号
}

// 创建配置 请求结构
type CreateConfigRequest struct {
	Capability  string `json:"capability" binding:"required"`  // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required"`        // 配置名称
	Description string `json:"description,omitempty"`          // 配置描述
	Platform    string `json:"platform" binding:"required"`    // 配置平台
	ConfigJSON  string `json:"config_json" binding:"required"` // 配置平台密钥等敏感信息（JSON）
}

// 更新配置 请求结构
type UpdateConfigRequest struct {
	GetConfigRequest
	Capability  string `json:"capability" binding:"required"`  // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required"`        // 配置名称
	Description string `json:"description,omitempty"`          // 配置描述
	ConfigJSON  string `json:"config_json" binding:"required"` // 配置平台密钥等敏感信息（JSON）
}

// 删除配置 请求结构
type DeleteConfigRequest struct {
	GetConfigRequest
}
