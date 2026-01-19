package types

// 日志页面 请求结构
type LogContentRequest struct {
	CommonPageRequest
	Type       string `json:"type,omitempty" form:"type,omitempty" default:"http_log"` // 日志类型
	Hostname   string `json:"hostname,omitempty" form:"hostname,omitempty"`            // 主机名
	Method     string `json:"method,omitempty" form:"method,omitempty"`                // 请求方法
	StatusCode int    `json:"status_code,omitempty" form:"status_code,omitempty"`      // 状态码
	FileName   string `json:"file_name,omitempty" form:"file_name,omitempty"`          // 文件名
}

// 日志页面 响应结构
type LogContentResponse struct {
	CommonPageResponse[any, LogContentRequest] // 操作日志列表
}
