package types

// 通知页面 请求结构
type WebhookRequest struct {
	CommonPageRequest
	ID       int64        `json:"id,omitempty" form:"id,omitempty"`               // 通知序号
	Name     string       `json:"name,omitempty" form:"name,omitempty"`           // 通知名称
	Platform string       `json:"platform,omitempty" form:"platform,omitempty"`   // 通知平台类型
	IsActive NullableBool `json:"is_active,omitempty" form:"is_active,omitempty"` // 是否活跃
}

// 通知页面 响应结构
type WebhookResponse struct {
	CommonPageResponse[GetWebhookResponse, WebhookRequest] // 通知列表
}

// 通知模板数据结构
type TemplateData struct {
	EventTitle     string `json:"event_title,omitempty"`      // 事件标题
	EventType      string `json:"event_type,omitempty"`       // 事件类型
	EventTime      string `json:"event_time,omitempty"`       // 事件时间
	EventBody      string `json:"event_body,omitempty"`       // 事件内容
	DeviceCode     string `json:"device_code,omitempty"`      // 设备编码 Device->Code
	DeviceName     string `json:"device_name,omitempty"`      // 设备名称 Device->Name
	IPv4Addr       string `json:"ipv4_addr,omitempty"`        // IP地址V4 Device->IPv4
	IPv6Addr       string `json:"ipv6_addr,omitempty"`        // IP地址V6 Device->IPv6
	RecordDomain   string `json:"record_domain,omitempty"`    // 记录域名 CDNRecord/DNSRecord->Domain
	RecordType     string `json:"record_type,omitempty"`      // 记录类型 CDNRecord/DNSRecord->Type
	RecordOldValue string `json:"record_old_value,omitempty"` // 旧记录值 CDNRecord/DNSRecord->Record
	RecordNewValue string `json:"record_new_value,omitempty"` // 新记录值 CDNRecord/DNSRecord->Record
	PlatformName   string `json:"platform_name,omitempty"`    // 平台名称 PlatformConfig->Name
	PlatformType   string `json:"platform_type,omitempty"`    // 平台类型 PlatformConfig->Platform
}
