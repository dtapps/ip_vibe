package types

// 通知页面 请求结构
type WebhookRequest struct {
	CommonPageRequest
	ID       int64        `json:"id,omitempty" form:"id,omitempty"`             // 通知序号
	Name     string       `json:"name,omitempty" form:"name,omitempty"`         // 通知名称
	Platform string       `json:"platform,omitempty" form:"platform,omitempty"` // 通知平台类型
	IsActive NullableBool `json:"is_active" form:"is_active,omitempty"`         // 是否活跃
}

// 通知页面 响应结构
type WebhookResponse struct {
	CommonPageResponse[GetWebhookResponse, WebhookRequest] // 通知列表
}
