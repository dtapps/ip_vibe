package types

// 获取全部通知订阅 请求结构
type GetWebhookSubsRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"title=是否活跃,description=指定是否只返回活跃的通知订阅,example=true"` // 是否活跃
}

// 获取通知订阅 响应结构
type GetWebhookSubResponse struct {
	ID          int64  `json:"id" jsonschema:"title=通知订阅序号,description=通知订阅的唯一标识符"`                    // 通知订阅序号
	Name        string `json:"name" jsonschema:"title=通知订阅名称,description=通知订阅的名称"`                     // 通知订阅名称
	Description string `json:"description,omitempty" jsonschema:"title=通知订阅描述,description=通知订阅的描述信息"`  // 通知订阅描述
	WebhookID   int64  `json:"webhook_id" jsonschema:"title=通知序号,description=关联的通知序号"`                 // 通知序号
	WebhookName string `json:"webhook_name,omitempty" jsonschema:"title=通知渠道名称,description=关联的通知渠道名称"` // 通知渠道名称
	ObjType     string `json:"obj_type" jsonschema:"title=订阅对象类型,description=订阅对象的类型"`                 // 订阅对象类型
	ObjID       int64  `json:"obj_id" jsonschema:"title=订阅对象主键,description=订阅对象的主键ID"`                 // 订阅对象主键
	ObjName     string `json:"obj_name,omitempty" jsonschema:"title=订阅对象名称,description=订阅对象的名称"`       // 订阅对象名称
	IsActive    bool   `json:"is_active" jsonschema:"title=是否启用,description=是否启用该通知订阅,default=true"`   // 是否启用
}

// 获取通知订阅 请求结构
type GetWebhookSubRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=通知订阅序号,description=通知订阅的唯一标识符,example=1"` // 通知订阅序号
}

// 创建通知订阅 请求结构
type CreateWebhookSubRequest struct {
	Name        string  `json:"name" binding:"required" jsonschema:"title=通知订阅名称,description=通知订阅的名称,example=CDN记录变更通知"`        // 通知订阅名称
	Description string  `json:"description,omitempty" jsonschema:"title=通知订阅描述,description=通知订阅的描述信息,example=用于接收CDN记录变更通知"`    // 通知订阅描述
	WebhookID   int64   `json:"webhook_id" binding:"required" jsonschema:"title=通知序号,description=关联的通知序号,example=1"`            // 通知序号
	ObjType     string  `json:"obj_type" binding:"required" jsonschema:"title=订阅对象类型,description=订阅对象的类型,example=cdn_record"`   // 订阅对象类型
	ObjIDs      []int64 `json:"obj_ids" binding:"required" jsonschema:"title=订阅对象主键列表,description=订阅对象的主键ID列表,example=[1,2,3]"` // 订阅对象主键列表
}

// 更新通知订阅 请求结构
type UpdateWebhookSubRequest struct {
	GetWebhookSubRequest
	Name        string `json:"name" binding:"required" jsonschema:"title=通知订阅名称,description=通知订阅的名称,example=更新后的通知订阅名称"` // 通知订阅名称
	Description string `json:"description,omitempty" jsonschema:"title=通知订阅描述,description=通知订阅的描述信息,example=更新后的通知订阅描述"` // 通知订阅描述
}

// 更新通知订阅状态 请求结构
type UpdateWebhookSubActiveRequest struct {
	GetWebhookSubRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"title=是否启用,description=是否启用该通知订阅,default=true,example=false"` // 是否启用
}

// 删除通知订阅 请求结构
type DeleteWebhookSubRequest struct {
	GetWebhookSubRequest
}
