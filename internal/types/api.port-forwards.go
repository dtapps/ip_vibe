package types

// 获取全部端口转发 请求结构
type GetPortForwardsRequest struct {
	IsActive *bool `json:"is_active" query:"is_active" binding:"required" jsonschema:"指定是否只返回活跃的端口转发记录"` // 是否活跃
}

// 获取端口转发 响应结构
type GetPortForwardResponse struct {
	ID          int64    `json:"id" jsonschema:"端口转发记录的唯一标识符"`                             // 设备端口转发序号
	Name        string   `json:"name" jsonschema:"端口转发的名称"`                                // 设备端口转发名称
	Description string   `json:"description,omitempty" jsonschema:"端口转发的描述信息"`             // 设备端口转发描述
	LocalPort   int      `json:"local_port" jsonschema:"本地监听的端口号"`                         // 本地监听端口
	RemoteHost  string   `json:"remote_host" jsonschema:"远程目标主机地址"`                        // 远程目标地址
	RemotePort  int      `json:"remote_port" jsonschema:"远程目标端口号"`                         // 远程目标端口
	Protocols   []string `json:"protocols" jsonschema:"支持的协议列表，支持 tcp4, tcp6, udp4, udp6"` // 协议列表，支持 tcp4, tcp6, udp4, udp6
	AutoForward bool     `json:"auto_forward" jsonschema:"是否启用防火墙自动放行"`                    // 防火墙自动放行
	IsActive    bool     `json:"is_active" jsonschema:"端口转发是否处于活跃状态"`                      // 是否活跃

	DeviceID   int64  `json:"device_id,omitempty" jsonschema:"关联的设备ID"`   // 设备ID
	DeviceName string `json:"device_name,omitempty" jsonschema:"关联的设备名称"` // 设备名称
}

// 获取端口转发 请求结构
type GetPortForwardRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"端口转发记录的唯一标识符"` // 设备端口转发序号
}

// 创建端口转发 请求结构
type CreatePortForwardRequest struct {
	Name        string   `json:"name" binding:"required" jsonschema:"端口转发的名称"`                                // 设备端口转发名称
	Description string   `json:"description,omitempty" jsonschema:"端口转发的描述信息"`                                // 设备端口转发描述
	LocalPort   int      `json:"local_port" binding:"required" jsonschema:"本地监听的端口号"`                         // 本地监听端口
	RemoteHost  string   `json:"remote_host" binding:"required" jsonschema:"远程目标主机地址"`                        // 远程目标地址
	RemotePort  int      `json:"remote_port" binding:"required" jsonschema:"远程目标端口号"`                         // 远程目标端口
	Protocols   []string `json:"protocols" binding:"required" jsonschema:"支持的协议列表，支持 tcp4, tcp6, udp4, udp6"` // 协议列表，支持 tcp4, tcp6, udp4, udp6
	DeviceID    int64    `json:"device_id" binding:"required" jsonschema:"关联的设备ID"`                           // 设备ID
}

// 更新端口转发 请求结构
type UpdatePortForwardRequest struct {
	GetPortForwardRequest
	Name        string   `json:"name" binding:"required" jsonschema:"端口转发的名称"`                                // 设备端口转发名称
	Description string   `json:"description,omitempty" jsonschema:"端口转发的描述信息"`                                // 设备端口转发描述
	LocalPort   int      `json:"local_port" binding:"required" jsonschema:"本地监听的端口号"`                         // 本地监听端口
	RemoteHost  string   `json:"remote_host" binding:"required" jsonschema:"远程目标主机地址"`                        // 远程目标地址
	RemotePort  int      `json:"remote_port" binding:"required" jsonschema:"远程目标端口号"`                         // 远程目标端口
	Protocols   []string `json:"protocols" binding:"required" jsonschema:"支持的协议列表，支持 tcp4, tcp6, udp4, udp6"` // 协议列表，支持 tcp4, tcp6, udp4, udp6
	DeviceID    int64    `json:"device_id" binding:"required" jsonschema:"关联的设备ID"`                           // 设备ID
}

// 更新端口转发防火墙 请求结构
type UpdatePortForwardForwardRequest struct {
	GetPortForwardRequest
	AutoForward bool `json:"auto_forward" jsonschema:"是否启用防火墙自动放行"` // 防火墙自动放行
}

// 更新端口转发状态 请求结构
type UpdatePortForwardActiveRequest struct {
	GetPortForwardRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃该端口转发记录"` // 是否活跃
}

// 删除端口转发 请求结构
type DeletePortForwardRequest struct {
	GetPortForwardRequest
}

// 检查端口转发健康状态 请求结构
type HealthCheckPortForwardRequest struct {
	GetPortForwardRequest
}
