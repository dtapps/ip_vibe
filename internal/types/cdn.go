package types

// CDN页面 请求结构
type CdnRequest struct {
	CommonPageRequest
}

// CDN页面 响应结构
type CdnResponse struct {
	CommonPageResponse[GetCdnRecordResponse] // CDN记录列表
}
