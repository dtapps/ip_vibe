package types

// 获取全部通知订阅 请求结构
type GetWebhookSubsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"指定是否只返回活跃的通知订阅"` // 是否活跃
}

// 获取通知订阅 响应结构
type GetWebhookSubResponse struct {
	ID          int64  `json:"id" jsonschema:"通知订阅的唯一标识符"`                    // 通知订阅序号
	Name        string `json:"name" jsonschema:"通知订阅的名称"`                     // 通知订阅名称
	Description string `json:"description,omitempty" jsonschema:"通知订阅的描述信息"`  // 通知订阅描述
	WebhookID   int64  `json:"webhook_id" jsonschema:"关联的通知序号"`               // 通知序号
	WebhookName string `json:"webhook_name,omitempty" jsonschema:"关联的通知渠道名称"` // 通知渠道名称
	ObjType     string `json:"obj_type" jsonschema:"订阅对象的类型"`                 // 订阅对象类型
	ObjID       int64  `json:"obj_id" jsonschema:"订阅对象的主键ID"`                 // 订阅对象主键
	ObjName     string `json:"obj_name,omitempty" jsonschema:"订阅对象的名称"`       // 订阅对象名称
	IsActive    bool   `json:"is_active" jsonschema:"是否启用该通知订阅"`              // 是否启用
}

// 获取通知订阅 请求结构
type GetWebhookSubRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"通知订阅的唯一标识符"` // 通知订阅序号
}

// 创建通知订阅 请求结构
type CreateWebhookSubRequest struct {
	Name        string  `json:"name" binding:"required" jsonschema:"通知订阅的名称"`        // 通知订阅名称
	Description string  `json:"description,omitempty" jsonschema:"通知订阅的描述信息"`        // 通知订阅描述
	WebhookID   int64   `json:"webhook_id" binding:"required" jsonschema:"关联的通知序号"`  // 通知序号
	ObjType     string  `json:"obj_type" binding:"required" jsonschema:"订阅对象的类型"`    // 订阅对象类型
	ObjIDs      []int64 `json:"obj_ids" binding:"required" jsonschema:"订阅对象的主键ID列表"` // 订阅对象主键列表
}

// 创建通知订阅 响应结构
type CreateWebhookSubResponse struct {
	IDs []int64 `json:"ids" jsonschema:"通知订阅的唯一标识符列表"` // 通知订阅序号列表
}

// 更新通知订阅 请求结构
type UpdateWebhookSubRequest struct {
	GetWebhookSubRequest
	Name        string `json:"name" binding:"required" jsonschema:"通知订阅的名称"` // 通知订阅名称
	Description string `json:"description,omitempty" jsonschema:"通知订阅的描述信息"` // 通知订阅描述
}

// 更新通知订阅状态 请求结构
type UpdateWebhookSubActiveRequest struct {
	GetWebhookSubRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否启用该通知订阅"` // 是否启用
}

// 删除通知订阅 请求结构
type DeleteWebhookSubRequest struct {
	GetWebhookSubRequest
}
