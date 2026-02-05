package types

// DNS页面 请求结构
type DnsRequest struct {
	CommonPageRequest
	ID          int64        `json:"id,omitempty" form:"id,omitempty"`                       // DNS记录序号
	ConfigID    int64        `json:"config_id,omitempty" form:"config_id,omitempty"`         // 平台配置序号
	RecordID    string       `json:"record_id,omitempty" form:"record_id,omitempty"`         // 平台侧记录ID
	SubDomain   string       `json:"sub_domain,omitempty" form:"sub_domain,omitempty"`       // 子域名（例如 @, www, blog, api）
	Type        string       `json:"type,omitempty" form:"type,omitempty"`                   // 记录类型（A、AAAA、CNAME）
	IPGroupID   int64        `json:"ip_group_id,omitempty" form:"ip_group_id,omitempty"`     // IP组序号（与 cdn_record_id 二选一）
	CdnRecordID int64        `json:"cdn_record_id,omitempty" form:"cdn_record_id,omitempty"` // CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）
	IsActive    NullableBool `json:"is_active,omitempty" form:"is_active,omitempty"`         // 是否活跃
}

// DNS页面 响应结构
type DnsResponse struct {
	CommonPageResponse[GetDnsRecordResponse, DnsRequest] // DNS记录列表
}
