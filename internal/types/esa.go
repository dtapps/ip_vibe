package types

// ESA页面 请求结构
type ESARequest struct {
	CommonPageRequest
}

// ESA页面 响应结构
type ESAResponse struct {
	CommonPageResponse[GetESAResponse, ESARequest]
}
