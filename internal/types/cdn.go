package types

// CDN页面 请求结构
type CdnRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"20"` // 每页数量
}

// CDN页面 响应结构
type CdnResponse struct {
	List     []GetCdnRecordsResponse `json:"list"`      // CDN 记录列表
	Page     int                     `json:"page"`      // 当前页码
	PageSize int                     `json:"page_size"` // 每页数量
	Total    int64                   `json:"total"`     // 总数量
}
