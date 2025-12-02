package types

// 日志页面 请求结构
type LogRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"20"` // 每页数量
}

// 日志页面 响应结构
type LogResponse struct {
	List     []GetLogResponse `json:"list"`      // 操作日志列表
	Page     int              `json:"page"`      // 当前页码
	PageSize int              `json:"page_size"` // 每页数量
	Total    int64            `json:"total"`     // 总数量
}
