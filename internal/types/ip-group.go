package types

type IPGroupResponse struct {
	List     []GetIPGroupsResponse `json:"list"`      // IP组列表
	Page     int                   `json:"page"`      // 当前页码
	PageSize int                   `json:"page_size"` // 每页数量
	Total    int                   `json:"total"`     // 总数量
}
