package types

// 设备页面 请求结构
type DeviceRequest struct {
	CommonPageRequest
}

// 设备页面 响应结构
type DeviceResponse struct {
	CommonResponseWithRequest             // 请求对象
	CommonPageResponse[GetDeviceResponse] // 设备列表
}
