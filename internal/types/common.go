package types

// 公共 页面分页 请求结构
type CommonPageRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 公共 页面分页 响应结构
type CommonPageResponse[T any] struct {
	List     []T   `json:"list"`      // 通知列表
	Page     int   `json:"page"`      // 当前页码
	PageSize int   `json:"page_size"` // 每页数量
	Total    int64 `json:"total"`     // 总数量
}
