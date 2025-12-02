package types

// 获取全部通知订阅 响应结构
type GetWebhookSubResponse struct {
	ID          int64  `json:"id"`                     // 通知订阅序号
	Name        string `json:"name"`                   // 通知订阅名称
	Description string `json:"description,omitempty"`  // 通知订阅描述
	WebhookID   int64  `json:"webhook_id"`             // 通知序号
	WebhookName string `json:"webhook_name,omitempty"` // 通知渠道名称
	ObjType     string `json:"obj_type"`               // 订阅对象类型
	ObjID       int64  `json:"obj_id"`                 // 订阅对象主键
	ObjName     string `json:"obj_name,omitempty"`     // 订阅对象名称
	IsActive    bool   `json:"is_active"`              // 是否启用
}

// 获取通知订阅 请求结构
type GetWebhookSubRequest struct {
	ID int64 `uri:"id" binding:"required"` // 通知订阅序号
}

// 创建通知订阅 请求结构
type CreateWebhookSubRequest struct {
	Name        string  `json:"name" binding:"required"`       // 通知订阅名称
	Description string  `json:"description,omitempty"`         // 通知订阅描述
	WebhookID   int64   `json:"webhook_id" binding:"required"` // 通知序号
	ObjType     string  `json:"obj_type" binding:"required"`   // 订阅对象类型
	ObjIDs      []int64 `json:"obj_ids" binding:"required"`    // 订阅对象主键列表
}

// 更新通知订阅 请求结构
type UpdateWebhookSubRequest struct {
	GetWebhookSubRequest
	Name        string `json:"name" binding:"required"` // 通知订阅名称
	Description string `json:"description,omitempty"`   // 通知订阅描述
}

// 更新通知订阅状态 请求结构
type UpdateWebhookSubActiveRequest struct {
	GetWebhookSubRequest
	IsActive bool `json:"is_active,omitempty"` // 是否启用
}

// 删除通知订阅 请求结构
type DeleteWebhookSubRequest struct {
	GetWebhookSubRequest
}
