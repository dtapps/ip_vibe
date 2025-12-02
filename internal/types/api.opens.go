package types

// 设备上报IP 请求结构
type DeviceReportDeviceIpRequest struct {
	DevicePortForwardsRequest
	DeviceName string `json:"device_name,omitempty"` // 设备名称
	IPV4Addr   string `json:"ipv4_addr,omitempty"`   // IP地址V4
	IPV6Addr   string `json:"ipv6_addr,omitempty"`   // IP地址V6
}

// 获取设备端口转发 请求结构
type DevicePortForwardsRequest struct {
	DeviceCode string `uri:"device_code" json:"-" binding:"required"` // 设备编码
}
