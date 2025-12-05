package types

// 端口转发页面 请求结构
type PortForwardRequest struct {
	CommonPageRequest
}

// 端口转发页面 响应结构
type PortForwardResponse struct {
	CommonPageResponse[GetPortForwardResponse] // 端口转发列表
}
