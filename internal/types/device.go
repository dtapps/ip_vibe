package types

// 设备页面 请求结构
type DeviceRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"20"` // 每页数量
}

// 设备页面 响应结构
type DeviceResponse struct {
	List     []GetDevicesResponse `json:"list"`      // 设备列表
	Page     int                  `json:"page"`      // 当前页码
	PageSize int                  `json:"page_size"` // 每页数量
	Total    int64                `json:"total"`     // 总数量
}
