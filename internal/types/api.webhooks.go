package types

// 获取全部通知 请求结构
type GetWebhooksRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"指定是否只返回活跃的通知"` // 是否活跃
}

// 获取通知 响应结构
type GetWebhookResponse struct {
	ID              int64  `json:"id" jsonschema:"通知的唯一标识符"`                       // 通知序号
	Name            string `json:"name" jsonschema:"通知的名称"`                        // 通知名称
	Description     string `json:"description" jsonschema:"通知的描述信息"`               // 通知描述
	Platform        string `json:"platform" jsonschema:"通知平台类型"`                   // 通知平台类型
	URL             string `json:"url" jsonschema:"Webhook的URL地址"`                 // Webhook URL
	Secret          string `json:"secret" jsonschema:"Webhook签名密钥或额外Token"`        // Webhook 签名密钥或额外Token
	TemplateTitle   string `json:"template_title" jsonschema:"通知标题的Go Template"`   // 通知标题 Go Template
	TemplateContent string `json:"template_content" jsonschema:"通知内容的Go Template"` // 通知内容 Go Template
	IsActive        bool   `json:"is_active" jsonschema:"是否启用该通知"`                 // 是否启用
}

// 获取通知 请求结构
type GetWebhookRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"通知的唯一标识符"` // 通知序号
}

// 创建通知 请求结构
type CreateWebhookRequest struct {
	Name            string `json:"name" binding:"required" jsonschema:"通知的名称"`                        // 通知名称
	Description     string `json:"description,omitempty" jsonschema:"通知的描述信息"`                        // 通知描述
	Platform        string `json:"platform" binding:"required" jsonschema:"通知平台类型"`                   // 通知平台类型
	URL             string `json:"url" jsonschema:"Webhook的URL地址"`                                    // Webhook URL
	Secret          string `json:"secret" jsonschema:"Webhook签名密钥或额外Token"`                           // Webhook 签名密钥或额外Token
	TemplateTitle   string `json:"template_title" binding:"required" jsonschema:"通知标题的Go Template"`   // 通知标题 Go Template
	TemplateContent string `json:"template_content" binding:"required" jsonschema:"通知内容的Go Template"` // 通知内容 Go Template
}

// 更新通知 请求结构
type UpdateWebhookRequest struct {
	GetWebhookRequest
	Name            string `json:"name" binding:"required" jsonschema:"通知的名称"`                        // 通知名称
	Description     string `json:"description,omitempty" jsonschema:"通知的描述信息"`                        // 通知描述
	URL             string `json:"url" jsonschema:"Webhook的URL地址"`                                    // Webhook URL
	Secret          string `json:"secret" jsonschema:"Webhook签名密钥或额外Token"`                           // Webhook 签名密钥或额外Token
	TemplateTitle   string `json:"template_title" binding:"required" jsonschema:"通知标题的Go Template"`   // 通知标题 Go Template
	TemplateContent string `json:"template_content" binding:"required" jsonschema:"通知内容的Go Template"` // 通知内容 Go Template
}

// 更新通知状态 请求结构
type UpdateWebhookActiveRequest struct {
	GetWebhookRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否启用该通知"` // 是否启用
}

// 删除通知 请求结构
type DeleteWebhookRequest struct {
	GetWebhookRequest
}

// 批量删除通知 请求结构
type DeleteWebhookBatchRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"通知的唯一标识符列表"` // 通知序号列表
}

// 测试通知 请求结构
type TestWebhookRequest struct {
	GetWebhookRequest
}
