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
	EventTitle     string // 事件标题
	EventType      string // 事件类型
	EventTime      string // 事件时间
	EventBody      string // 事件内容
	DeviceCode     string // 设备编码 Device->Code
	DeviceName     string // 设备名称 Device->Name
	IPv4Addr       string // IP地址V4 Device->IPv4
	IPv6Addr       string // IP地址V6 Device->IPv6
	RecordDomain   string // 记录域名 CDNRecord/DNSRecord->Domain
	RecordType     string // 记录类型 CDNRecord/DNSRecord->Type
	RecordOldValue string // 旧记录值 CDNRecord/DNSRecord->Record
	RecordNewValue string // 新记录值 CDNRecord/DNSRecord->Record
	PlatformName   string // 平台名称 PlatformConfig->Name
	PlatformType   string // 平台类型 PlatformConfig->Platform
}
