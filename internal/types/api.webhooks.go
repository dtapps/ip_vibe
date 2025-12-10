package types

// 获取全部通知 请求结构
type GetWebhooksRequest struct {
	IsActive *bool `json:"is_active" binding:"required"` // 是否活跃
}

// 获取通知 响应结构
type GetWebhookResponse struct {
	ID              int64    `json:"id"`                         // 通知序号
	Name            string   `json:"name"`                       // 通知名称
	Description     string   `json:"description,omitempty"`      // 通知描述
	Platform        string   `json:"platform"`                   // 通知平台类型
	URL             string   `json:"url"`                        // Webhook URL
	Secret          string   `json:"secret"`                     // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                     // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title,omitempty"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content,omitempty"` // 通知内容 Go Template
	IsActive        bool     `json:"is_active"`                  // 是否启用
}

// 获取通知 请求结构
type GetWebhookRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 通知序号
}

// 创建通知 请求结构
type CreateWebhookRequest struct {
	Name            string   `json:"name" binding:"required"`             // 通知名称
	Description     string   `json:"description,omitempty"`               // 通知描述
	Platform        string   `json:"platform" binding:"required"`         // 通知平台类型
	URL             string   `json:"url"`                                 // Webhook URL
	Secret          string   `json:"secret"`                              // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                              // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title" binding:"required"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content" binding:"required"` // 通知内容 Go Template
}

// 更新通知 请求结构
type UpdateWebhookRequest struct {
	GetWebhookRequest
	Name            string   `json:"name" binding:"required"`             // 通知名称
	Description     string   `json:"description,omitempty"`               // 通知描述
	Secret          string   `json:"secret"`                              // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                              // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title" binding:"required"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content" binding:"required"` // 通知内容 Go Template
}

// 更新通知状态 请求结构
type UpdateWebhookActiveRequest struct {
	GetWebhookRequest
	IsActive bool `json:"is_active,omitempty"` // 是否启用
}

// 删除通知 请求结构
type DeleteWebhookRequest struct {
	GetWebhookRequest
}
