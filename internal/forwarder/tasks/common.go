package tasks

// Config 配置
type Config struct {
	DeviceCode string `json:"device_code"` // 设备编号
	DeviceName string `json:"device_name"` // 设备名称
}

type PortForwardTaskApiList struct {
	ID         string   // 设备端口转发序号
	LocalPort  int      // 本地监听端口
	RemoteHost string   // 远程目标地址
	RemotePort int      // 远程目标端口
	Protocols  []string // 协议列表，支持 tcp4, tcp6, udp4, udp6
}
