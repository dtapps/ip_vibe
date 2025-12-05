package types

// 平台配置页面 请求结构
type ConfigRequest struct {
	CommonPageRequest
}

// 平台配置页面 响应结构
type ConfigResponse struct {
	CommonPageResponse[GetConfigResponse] // 配置列表
}
