package types

// 平台配置页面 请求结构
type PlatformConfigRequest struct {
	CommonPageRequest
	ID         int64  `json:"id,omitempty" form:"id,omitempty"`                 // 平台配置序号
	Capability string `json:"capability,omitempty" form:"capability,omitempty"` // 平台支持的服务能力：all（全部）dns（仅DNS）、cdn（仅CDN）、ssl（仅SSL）
	Name       string `json:"name,omitempty" form:"name,omitempty"`             // 平台配置名称
	Platform   string `json:"platform,omitempty" form:"platform,omitempty"`     // 平台类型
}

// 平台配置页面 响应结构
type PlatformConfigResponse struct {
	CommonPageResponse[GetPlatformConfigResponse, PlatformConfigRequest] // 配置列表
}
