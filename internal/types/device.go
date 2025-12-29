package types

// 设备页面 请求结构
type DeviceRequest struct {
	CommonPageRequest
}

// 设备页面 响应结构
type DeviceResponse struct {
	CommonPageResponse[GetDeviceResponse] // 设备列表
}

// 设备端口转发统计 请求结构
type DevicePortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"设备序号"`
}
