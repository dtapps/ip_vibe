package types

type LogResponse struct {
	List     []GetLogsResponse `json:"list"`      // 操作日志列表
	Page     int               `json:"page"`      // 当前页码
	PageSize int               `json:"page_size"` // 每页数量
	Total    int               `json:"total"`     // 总数量
}
