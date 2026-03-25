package types

// 通知订阅页面 请求结构
type WebhookSubRequest struct {
	CommonPageRequest
	ID        int64        `json:"id,omitempty" form:"id,omitempty"`                 // 通知订阅序号
	Name      string       `json:"name,omitempty" form:"name,omitempty"`             // 通知订阅名称
	WebhookID int64        `json:"webhook_id,omitempty" form:"webhook_id,omitempty"` // 通知序号
	ObjType   string       `json:"obj_type,omitempty" form:"obj_type,omitempty"`     // 对象类型
	ObjID     int64        `json:"obj_id,omitempty" form:"obj_id,omitempty"`         // 对象主键
	IsActive  NullableBool `json:"is_active" form:"is_active,omitempty"`             // 是否活跃
}

// 通知订阅页面 响应结构
type WebhookSubResponse struct {
	CommonPageResponse[GetWebhookSubResponse, WebhookSubRequest] // 通知订阅列表
}
