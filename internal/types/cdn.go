package types

// CDN页面 请求结构
type CdnRequest struct {
	CommonPageRequest
	ID        int64        `json:"id,omitempty" form:"id,omitempty"`                   // CDN记录序号
	ConfigID  int64        `json:"config_id,omitempty" form:"config_id,omitempty"`     // 平台配置序号
	RecordID  string       `json:"record_id,omitempty" form:"record_id,omitempty"`     // 平台侧记录ID
	SubDomain string       `json:"sub_domain,omitempty" form:"sub_domain,omitempty"`   // 子域名（例如 @, www, blog, api）
	Type      string       `json:"type,omitempty" form:"type,omitempty"`               // 记录类型（A、AAAA）
	IPGroupID int64        `json:"ip_group_id,omitempty" form:"ip_group_id,omitempty"` // IP组序号
	IsActive  NullableBool `json:"is_active,omitempty" form:"is_active,omitempty"`     // 是否活跃
}

// CDN页面 响应结构
type CdnResponse struct {
	CommonPageResponse[GetCdnRecordResponse, CdnRequest] // CDN记录列表
}
