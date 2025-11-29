package types

// 设备上报IP 请求结构
type DeviceReportDeviceIpRequest struct {
	DeviceName string `json:"device_name,omitempty"` // 设备名称
	IPV4Addr   string `json:"ipv4_addr,omitempty"`   // IP地址V4
	IPV6Addr   string `json:"ipv6_addr,omitempty"`   // IP地址V6
}
