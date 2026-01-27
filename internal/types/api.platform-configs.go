package types

import (
	"encoding/json"
	"time"
)

// 获取全部平台配置 请求结构
type GetPlatformConfigsRequest struct {
}

// 获取平台配置 响应结构
type GetPlatformConfigResponse struct {
	ID          int64           `json:"id" jsonschema:"配置序号"`                                                          // 配置序号
	Capability  string          `json:"capability" jsonschema:"平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）"` // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string          `json:"name" jsonschema:"配置名称"`                                                        // 配置名称
	Description string          `json:"description" jsonschema:"配置描述"`                                                 // 配置描述
	Platform    string          `json:"platform" jsonschema:"配置平台"`                                                    // 配置平台
	ConfigJSON  json.RawMessage `json:"config_json,omitempty" jsonschema:"平台密钥等敏感信息（JSON）"`                            // 平台密钥等敏感信息（JSON）

	ConfigJSONDomainName string `json:"-"` // 配置平台域名
	ConfigJSONEmail      string `json:"-"` // 配置平台邮箱

	CreatedAt time.Time `json:"-"` // 创建时间
	UpdatedAt time.Time `json:"-"` // 更新时间
}

// 获取平台配置 请求结构
type GetPlatformConfigRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"配置序号"` // 配置序号
}

// 创建平台配置 请求结构
type CreatePlatformConfigRequest struct {
	Capability  string `json:"capability" binding:"required" jsonschema:"平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）"` // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string `json:"name" binding:"required" jsonschema:"配置名称"`                                                        // 配置名称
	Description string `json:"description,omitempty" jsonschema:"配置描述"`                                                          // 配置描述
	Platform    string `json:"platform" binding:"required" jsonschema:"配置平台"`                                                    // 配置平台
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"配置平台密钥等敏感信息（JSON）"`                                    // 配置平台密钥等敏感信息（JSON）
}

// 更新平台配置 请求结构
type UpdatePlatformConfigRequest struct {
	GetPlatformConfigRequest
	Capability  string `json:"capability" binding:"required" jsonschema:"平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）"` // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string `json:"name" binding:"required" jsonschema:"配置名称"`                                                        // 配置名称
	Description string `json:"description,omitempty" jsonschema:"配置描述"`                                                          // 配置描述
	ConfigJSON  string `json:"config_json" binding:"required" jsonschema:"配置平台密钥等敏感信息（JSON）"`                                    // 配置平台密钥等敏感信息（JSON）
}

// 删除平台配置 请求结构
type DeletePlatformConfigRequest struct {
	GetPlatformConfigRequest
}

// 批量删除平台配置 请求结构
type BatchDeletePlatformConfigRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"配置序号列表"` // 配置序号列表
}
