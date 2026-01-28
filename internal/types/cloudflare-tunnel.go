package types

// CloudflareTunnel页面 请求结构
type CloudflareTunnelRequest struct {
	CommonPageRequest
}

// CloudflareTunnel页面 响应结构
type CloudflareTunnelResponse struct {
	CommonPageResponse[GetCloudflareTunnelResponse, CloudflareTunnelRequest] // TUN记录列表
}
