package types

// 日志页面 请求结构
type LogRequest struct {
	CommonPageRequest
}

// 日志页面 响应结构
type LogResponse struct {
	CommonResponseWithRequest          // 请求对象
	CommonPageResponse[GetLogResponse] // 操作日志列表
}
