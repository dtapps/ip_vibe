package types

// IP组页面 请求结构
type IPGroupRequest struct {
	CommonPageRequest
}

// IP组页面 响应结构
type IPGroupResponse struct {
	CommonPageResponse[GetIPGroupResponse, IPGroupRequest] // IP组列表
}
