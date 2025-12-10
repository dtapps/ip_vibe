package types

// 获取全部端口转发 请求结构
type GetPortForwardsRequest struct {
	IsActive *bool `json:"is_active" binding:"required"` // 是否活跃
}

// 获取端口转发 响应结构
type GetPortForwardResponse struct {
	ID          int64    `json:"id"`                    // 设备端口转发序号
	Name        string   `json:"name"`                  // 设备端口转发名称
	Description string   `json:"description,omitempty"` // 设备端口转发描述
	LocalPort   int      `json:"local_port"`            // 本地监听端口
	RemoteHost  string   `json:"remote_host"`           // 远程目标地址
	RemotePort  int      `json:"remote_port"`           // 远程目标端口
	Protocols   []string `json:"protocols"`             // 协议列表，支持 tcp4, tcp6, udp4, udp6
	AutoForward bool     `json:"auto_forward"`          // 防火墙自动放行
	IsActive    bool     `json:"is_active"`             // 是否活跃

	DeviceID   int64  `json:"device_id,omitempty"`   // 设备ID
	DeviceName string `json:"device_name,omitempty"` // 设备名称
}

// 获取端口转发 请求结构
type GetPortForwardRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 设备端口转发序号
}

// 创建端口转发 请求结构
type CreatePortForwardRequest struct {
	Name        string   `json:"name" binding:"required"`        // 设备端口转发名称
	Description string   `json:"description,omitempty"`          // 设备端口转发描述
	LocalPort   int      `json:"local_port" binding:"required"`  // 本地监听端口
	RemoteHost  string   `json:"remote_host" binding:"required"` // 远程目标地址
	RemotePort  int      `json:"remote_port" binding:"required"` // 远程目标端口
	Protocols   []string `json:"protocols" binding:"required"`   // 协议列表，支持 tcp4, tcp6, udp4, udp6
	DeviceID    int64    `json:"device_id" binding:"required"`   // 设备ID
}

// 更新端口转发 请求结构
type UpdatePortForwardRequest struct {
	GetPortForwardRequest
	Name        string   `json:"name" binding:"required"`        // 设备端口转发名称
	Description string   `json:"description,omitempty"`          // 设备端口转发描述
	LocalPort   int      `json:"local_port" binding:"required"`  // 本地监听端口
	RemoteHost  string   `json:"remote_host" binding:"required"` // 远程目标地址
	RemotePort  int      `json:"remote_port" binding:"required"` // 远程目标端口
	Protocols   []string `json:"protocols" binding:"required"`   // 协议列表，支持 tcp4, tcp6, udp4, udp6
	DeviceID    int64    `json:"device_id" binding:"required"`   // 设备ID
}

// 更新端口转发防火墙 请求结构
type UpdatePortForwardForwardRequest struct {
	GetPortForwardRequest
	AutoForward bool `json:"auto_forward"` // 防火墙自动放行
}

// 更新端口转发状态 请求结构
type UpdatePortForwardActiveRequest struct {
	GetPortForwardRequest
	IsActive bool `json:"is_active,omitempty"` // 是否活跃
}

// 删除端口转发 请求结构
type DeletePortForwardRequest struct {
	GetPortForwardRequest
}

// 检查端口转发健康状态 请求结构
type HealthcheckPortForwardRequest struct {
	GetPortForwardRequest
}
