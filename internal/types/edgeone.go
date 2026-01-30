package types

// EdgeOne页面 请求结构
type EdgeOneRequest struct {
	CommonPageRequest
}

// EdgeOne页面 响应结构
type EdgeOneResponse struct {
	CommonPageResponse[GetEdgeOneResponse, EdgeOneRequest] // TUN记录列表
}
