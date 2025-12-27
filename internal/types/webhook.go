package types

// 通知页面 请求结构
type WebhookRequest struct {
	CommonPageRequest
}

// 通知页面 响应结构
type WebhookResponse struct {
	CommonResponseWithRequest              // 请求对象
	CommonPageResponse[GetWebhookResponse] // 通知列表
}
