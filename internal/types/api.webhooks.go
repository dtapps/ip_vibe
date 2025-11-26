package types

// 获取全部通知 响应结构
type GetWebhooksResponse struct {
	ID              int64    `json:"id"`                         // 通知序号
	Name            string   `json:"name"`                       // 通知名称
	Platform        string   `json:"platform"`                   // 通知平台类型
	URL             string   `json:"url"`                        // Webhook URL
	Secret          string   `json:"secret"`                     // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                     // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title,omitempty"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content,omitempty"` // 通知内容 Go Template
	IsActive        bool     `json:"is_active"`                  // 是否启用
}

// 创建通知 请求结构
type CreateWebhookRequest struct {
	Name            string   `json:"name"`                       // 通知名称
	Platform        string   `json:"platform"`                   // 通知平台类型
	URL             string   `json:"url"`                        // Webhook URL
	Secret          string   `json:"secret"`                     // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                     // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title,omitempty"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content,omitempty"` // 通知内容 Go Template
}

// 更新通知 请求结构
type UpdateWebhookRequest struct {
	Name            string   `json:"name"`                       // 通知名称
	Secret          string   `json:"secret"`                     // Webhook 签名密钥或额外Token
	Events          []string `json:"events"`                     // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title,omitempty"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content,omitempty"` // 通知内容 Go Template
}

// 更新通知状态 请求结构
type UpdateWebhookActiveRequest struct {
	IsActive bool `json:"is_active,omitempty"` // 是否启用
}
