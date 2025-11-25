package types

// 获取全部通知订阅 响应结构
type GetWebhookSubscriptionsResponse struct {
	ID          int64  `json:"id"`                     // 通知订阅序号
	Name        string `json:"name"`                   // 通知订阅名称
	WebhookID   int64  `json:"webhook_id"`             // 通知序号
	WebhookName string `json:"webhook_name,omitempty"` // 通知渠道名称
	ObjType     string `json:"obj_type"`               // 订阅对象类型
	ObjID       int64  `json:"obj_id"`                 // 订阅对象主键
	IsActive    bool   `json:"is_active"`              // 是否启用
}

// 创建通知订阅 请求结构
type CreateWebhookSubscriptionRequest struct {
	Name      string `json:"name"`       // 通知订阅名称
	WebhookID int64  `json:"webhook_id"` // 通知序号
	ObjType   string `json:"obj_type"`   // 订阅对象类型
	ObjID     int64  `json:"obj_id"`     // 订阅对象主键
}

// 更新通知订阅 请求结构
type UpdateWebhookSubscriptionRequest struct {
	Name string `json:"name"` // 通知订阅名称
}

// 更新通知订阅状态 请求结构
type UpdateWebhookSubscriptionActiveRequest struct {
	IsActive bool `json:"is_active,omitempty"` // 是否启用
}
