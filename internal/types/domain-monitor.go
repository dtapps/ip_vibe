package types

// 域名监控页面 请求结构
type DomainMonitorRequest struct {
	CommonPageRequest
}

// 域名监控页面 响应结构
type DomainMonitorResponse struct {
	CommonPageResponse[GetDomainMonitorResponse] // 域名监控列表
}
