package types

type ConfigResponse struct {
	List     []GetConfigsResponse `json:"list"`      // 配置列表
	Page     int                  `json:"page"`      // 当前页码
	PageSize int                  `json:"page_size"` // 每页数量
	Total    int                  `json:"total"`     // 总数量
}
