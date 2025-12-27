package types

// IP组页面 请求结构
type IPGroupRequest struct {
	CommonPageRequest
}

// IP组页面 响应结构
type IPGroupResponse struct {
	CommonResponseWithRequest              // 请求对象
	CommonPageResponse[GetIPGroupResponse] // IP组列表
}
