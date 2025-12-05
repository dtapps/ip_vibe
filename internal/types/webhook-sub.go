package types

// 通知订阅页面 请求结构
type WebhookSubRequest struct {
	CommonPageRequest
}

// 通知订阅页面 响应结构
type WebhookSubResponse struct {
	CommonPageResponse[GetWebhookSubResponse] // 通知订阅列表
}
