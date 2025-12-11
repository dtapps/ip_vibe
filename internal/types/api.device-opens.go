package types

// 设备上报IP 请求结构
type DeviceReportDeviceIpRequest struct {
	DevicePortForwardsRequest
	DeviceName string `json:"device_name,omitempty"` // 设备名称
	IPV4Addr   string `json:"ipv4_addr,omitempty"`   // IP地址V4
	IPV6Addr   string `json:"ipv6_addr,omitempty"`   // IP地址V6
}

// 设备上报IP 响应结构
type DeviceReportDeviceIpResponse struct {
	DeviceCode string `json:"device_code,omitempty"` // 设备编码
	Msg        string `json:"msg,omitempty"`         // 消息
}

// 获取设备端口转发 请求结构
type DevicePortForwardsRequest struct {
	DeviceCode string `uri:"device_code" json:"-" binding:"required"` // 设备编码
}

// 获取设备端口转发 响应结构
type DevicePortForwardsResponse struct {
	ID          int64    `json:"id"`           // 设备端口转发序号
	LocalPort   int      `json:"local_port"`   // 本地监听端口
	RemoteHost  string   `json:"remote_host"`  // 远程目标地址
	RemotePort  int      `json:"remote_port"`  // 远程目标端口
	Protocols   []string `json:"protocols"`    // 协议列表，支持 tcp4, tcp6, udp4, udp6
	AutoForward bool     `json:"auto_forward"` // 防火墙自动放行
}
