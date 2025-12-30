package types

// DNS页面 请求结构
type DnsRequest struct {
	CommonPageRequest
}

// DNS页面 响应结构
type DnsResponse struct {
	CommonPageResponse[GetDnsRecordResponse, DnsRequest] // DNS记录列表
}
