package types

// IP组页面 请求结构
type IPGroupRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// IP组页面 响应结构
type IPGroupResponse struct {
	List     []GetIPGroupResponse `json:"list"`      // IP组列表
	Page     int                  `json:"page"`      // 当前页码
	PageSize int                  `json:"page_size"` // 每页数量
	Total    int64                `json:"total"`     // 总数量
}
