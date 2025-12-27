package types

// SSL页面 请求结构
type SSLRequest struct {
	CommonPageRequest
}

// SSL页面 响应结构
type SSLResponse struct {
	CommonResponseWithRequest                // 请求对象
	CommonPageResponse[GetSSLRecordResponse] // SSL记录列表
}
