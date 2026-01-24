package types

// SSL导入页面 请求结构
type SSLImportRequest struct {
	CommonPageRequest
}

// SSL导入页面 响应结构
type SSLImportResponse struct {
	CommonPageResponse[GetSSLImportResponse, SSLImportRequest] // SSL记录列表
}
