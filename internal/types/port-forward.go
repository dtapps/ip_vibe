package types

// 端口转发页面 请求结构
type PortForwardRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 端口转发页面 响应结构
type PortForwardResponse struct {
	List     []GetPortForwardResponse `json:"list"`      // 端口转发列表
	Page     int                      `json:"page"`      // 当前页码
	PageSize int                      `json:"page_size"` // 每页数量
	Total    int64                    `json:"total"`     // 总数量
}
