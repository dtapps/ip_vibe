package types

// DNS页面 请求结构
type DnsRequest struct {
	CommonPageRequest
}

// DNS页面 响应结构
type DnsResponse struct {
	CommonResponseWithRequest                // 请求对象
	CommonPageResponse[GetDnsRecordResponse] // DNS记录列表
}
