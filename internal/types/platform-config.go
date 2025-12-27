package types

// 平台配置页面 请求结构
type PlatformConfigRequest struct {
	CommonPageRequest
}

// 平台配置页面 响应结构
type PlatformConfigResponse struct {
	CommonResponseWithRequest                     // 请求对象
	CommonPageResponse[GetPlatformConfigResponse] // 配置列表
}
