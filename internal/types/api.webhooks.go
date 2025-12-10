package types

// 获取全部通知 请求结构
type GetWebhooksRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"title=是否活跃,description=指定是否只返回活跃的通知,example=true"` // 是否活跃
}

// 获取通知 响应结构
type GetWebhookResponse struct {
	ID              int64    `json:"id" jsonschema:"title=通知序号,description=通知的唯一标识符"`                                   // 通知序号
	Name            string   `json:"name" jsonschema:"title=通知名称,description=通知的名称"`                                    // 通知名称
	Description     string   `json:"description,omitempty" jsonschema:"title=通知描述,description=通知的描述信息"`                 // 通知描述
	Platform        string   `json:"platform" jsonschema:"title=通知平台类型,description=通知平台类型"`                             // 通知平台类型
	URL             string   `json:"url" jsonschema:"title=Webhook URL,description=Webhook的URL地址"`                      // Webhook URL
	Secret          string   `json:"secret" jsonschema:"title=签名密钥,description=Webhook签名密钥或额外Token"`                    // Webhook 签名密钥或额外Token
	Events          []string `json:"events" jsonschema:"title=事件类型列表,description=订阅的事件类型列表"`                            // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title,omitempty" jsonschema:"title=通知标题模板,description=通知标题的Go Template"`   // 通知标题 Go Template
	TemplateContent string   `json:"template_content,omitempty" jsonschema:"title=通知内容模板,description=通知内容的Go Template"` // 通知内容 Go Template
	IsActive        bool     `json:"is_active" jsonschema:"title=是否启用,description=是否启用该通知,default=true"`                // 是否启用
}

// 获取通知 请求结构
type GetWebhookRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=通知序号,description=通知的唯一标识符,example=1"` // 通知序号
}

// 创建通知 请求结构
type CreateWebhookRequest struct {
	Name            string   `json:"name" binding:"required" jsonschema:"title=通知名称,description=通知的名称,example=更新后的通知名称"`                                                                 // 通知名称
	Description     string   `json:"description,omitempty" jsonschema:"title=通知描述,description=通知的描述信息,example=用于接收系统告警"`                                                                 // 通知描述
	Platform        string   `json:"platform" binding:"required" jsonschema:"title=通知平台类型,description=通知平台类型,example=slack"`                                                             // 通知平台类型
	URL             string   `json:"url" jsonschema:"title=Webhook URL,description=Webhook的URL地址,example=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"` // Webhook URL
	Secret          string   `json:"secret" jsonschema:"title=签名密钥,description=Webhook签名密钥或额外Token,example=XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"`                                            // Webhook 签名密钥或额外Token
	Events          []string `json:"events" jsonschema:"title=事件类型列表,description=订阅的事件类型列表,example=[\"cdn_record_created\",\"cdn_record_updated\"]"`                                     // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title" binding:"required" jsonschema:"title=通知标题模板,description=通知标题的Go Template,example=系统告警"`                                              // 通知标题 Go Template
	TemplateContent string   `json:"template_content" binding:"required" jsonschema:"title=通知内容模板,description=通知内容的Go Template,example=检测到CDN记录变更，请及时处理。"`                               // 通知内容 Go Template
}

// 更新通知 请求结构
type UpdateWebhookRequest struct {
	GetWebhookRequest
	Name            string   `json:"name" binding:"required" jsonschema:"title=通知名称,description=通知的名称,example=更新后的通知名称"`                             // 通知名称
	Description     string   `json:"description,omitempty" jsonschema:"title=通知描述,description=通知的描述信息,example=更新后的通知描述"`                             // 通知描述
	Secret          string   `json:"secret" jsonschema:"title=签名密钥,description=Webhook签名密钥或额外Token,example=updated_secret_token"`                    // Webhook 签名密钥或额外Token
	Events          []string `json:"events" jsonschema:"title=事件类型列表,description=订阅的事件类型列表,example=[\"cdn_record_created\",\"cdn_record_deleted\"]"` // 订阅的事件类型列表
	TemplateTitle   string   `json:"template_title" binding:"required" jsonschema:"title=通知标题模板,description=通知标题的Go Template,example=更新后的标题模板"`      // 通知标题 Go Template
	TemplateContent string   `json:"template_content" binding:"required" jsonschema:"title=通知内容模板,description=通知内容的Go Template,example=更新后的内容模板"`    // 通知内容 Go Template
}

// 更新通知状态 请求结构
type UpdateWebhookActiveRequest struct {
	GetWebhookRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"title=是否启用,description=是否启用该通知,default=true,example=false"` // 是否启用
}

// 删除通知 请求结构
type DeleteWebhookRequest struct {
	GetWebhookRequest
}
