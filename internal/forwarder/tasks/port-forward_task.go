package tasks

// PortForwardTaskConfig 配置
type PortForwardTaskConfig struct {
	DeviceCode string `json:"device_code"` // 设备编号
	DeviceName string `json:"device_name"` // 设备名称

	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

type PortForwardTaskApiList struct {
	ID         string   // 设备端口转发序号
	LocalPort  int      // 本地监听端口
	RemoteHost string   // 远程目标地址
	RemotePort int      // 远程目标端口
	Protocols  []string // 协议列表，支持 tcp4, tcp6, udp4, udp6
}
