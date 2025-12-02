package types

// 通知页面 请求结构
type WebhookRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"20"` // 每页数量
}

// 通知页面 响应结构
type WebhookResponse struct {
	List     []GetWebhooksResponse `json:"list"`      // 通知列表
	Page     int                   `json:"page"`      // 当前页码
	PageSize int                   `json:"page_size"` // 每页数量
	Total    int64                 `json:"total"`     // 总数量
}
