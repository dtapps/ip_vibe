package types

// TUN页面 请求结构
type TunRequest struct {
	CommonPageRequest
	ID        int64        `json:"id,omitempty" form:"id,omitempty"`                 // TUN记录序号
	ConfigID  int64        `json:"config_id,omitempty" form:"config_id,omitempty"`   // 平台配置序号
	RecordID  string       `json:"record_id,omitempty" form:"record_id,omitempty"`   // 平台侧记录ID
	SubDomain string       `json:"sub_domain,omitempty" form:"sub_domain,omitempty"` // 子域名（如 www, blog, api）
	ServeType string       `json:"serve_type,omitempty" form:"serve_type,omitempty"` // 服务类型
	IsActive  NullableBool `json:"is_active,omitempty" form:"is_active,omitempty"`   // 是否活跃
}

// TUN页面 响应结构
type TunResponse struct {
	CommonPageResponse[GetTunRecordResponse, TunRequest] // TUN记录列表
}
