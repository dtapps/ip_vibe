package types

import "encoding/json"

// 获取配置 响应结构
type GetConfigResponse struct {
	ID          int64           `json:"id,omitempty" jsonschema:"配置的唯一标识符"`                                 // 配置序号
	Capability  string          `json:"capability" jsonschema:"平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）"` // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string          `json:"name" jsonschema:"配置的名称"`                                            // 配置名称
	Description string          `json:"description,omitempty" jsonschema:"配置的描述信息"`                         // 配置描述
	Platform    string          `json:"platform" jsonschema:"配置所属的平台"`                                      // 配置平台
	ConfigJSON  json.RawMessage `json:"config_json,omitempty" jsonschema:"配置平台密钥等敏感信息（JSON格式）"`             // 配置平台密钥等敏感信息（JSON）
}

// 获取配置 请求结构
type GetConfigRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"配置的唯一标识符"` // 配置序号
}

// 创建配置 请求结构
type CreateConfigRequest struct {
	Capability  string `json:"capability" binding:"required" jsonschema:"平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）"` // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required" jsonschema:"配置的名称"`                                            // 配置名称
	Description string `json:"description,omitempty" jsonschema:"配置的描述信息"`                                            // 配置描述
	Platform    string `json:"platform" binding:"required" jsonschema:"配置所属的平台"`                                      // 配置平台
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"配置平台密钥等敏感信息（JSON格式）"`                       // 配置平台密钥等敏感信息（JSON）
}

// 更新配置 请求结构
type UpdateConfigRequest struct {
	GetConfigRequest
	Capability  string `json:"capability" binding:"required" jsonschema:"平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）"` // 平台支持的服务能力：dns（仅DNS）、cdn（仅CDN）、all（DNS+CDN）
	Name        string `json:"name" binding:"required" jsonschema:"配置的名称"`                                            // 配置名称
	Description string `json:"description,omitempty" jsonschema:"配置的描述信息"`                                            // 配置描述
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"配置平台密钥等敏感信息（JSON格式）"`                       // 配置平台密钥等敏感信息（JSON）
}

// 删除配置 请求结构
type DeleteConfigRequest struct {
	GetConfigRequest
}
