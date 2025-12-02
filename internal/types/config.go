package types

// 平台配置页面 请求结构
type ConfigRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 平台配置页面 响应结构
type ConfigResponse struct {
	List     []GetConfigResponse `json:"list"`      // 配置列表
	Page     int                 `json:"page"`      // 当前页码
	PageSize int                 `json:"page_size"` // 每页数量
	Total    int64               `json:"total"`     // 总数量
}
