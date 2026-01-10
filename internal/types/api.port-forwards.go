package types

import "time"

// 获取全部端口转发 请求结构
type GetPortForwardsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"指定是否只返回活跃的端口转发记录"` // 是否活跃
}

// 获取端口转发 响应结构
type GetPortForwardResponse struct {
	ID          int64    `json:"id" jsonschema:"端口转发记录的唯一标识符"`                             // 设备端口转发序号
	Name        string   `json:"name" jsonschema:"端口转发的名称"`                                // 设备端口转发名称
	Description string   `json:"description" jsonschema:"端口转发的描述信息"`                       // 设备端口转发描述
	LocalPort   int      `json:"local_port" jsonschema:"本地监听的端口号"`                         // 本地监听端口
	RemoteHost  string   `json:"remote_host" jsonschema:"远程目标主机地址"`                        // 远程目标地址
	RemotePort  int      `json:"remote_port" jsonschema:"远程目标端口号"`                         // 远程目标端口
	Protocols   []string `json:"protocols" jsonschema:"支持的协议列表，支持 tcp4, tcp6, udp4, udp6"` // 协议列表，支持 tcp4, tcp6, udp4, udp6
	AutoForward bool     `json:"auto_forward" jsonschema:"是否启用防火墙自动放行"`                    // 防火墙自动放行
	IsActive    bool     `json:"is_active" jsonschema:"端口转发是否处于活跃状态"`                      // 是否活跃

	DeviceID   int64  `json:"device_id" jsonschema:"关联的设备ID"`   // 设备ID
	DeviceName string `json:"device_name" jsonschema:"关联的设备名称"` // 设备名称
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

// 获取端口转发统计数据 请求结构
type GetPortForwardStatsRequest struct {
	GetPortForwardRequest
	Range string `json:"range" form:"range" binding:"required" jsonschema:"查询时间范围，支持 1h, 6h, 24h, 7d"` // 查询时间范围，支持 1h, 6h, 24h, 7d
}

// 获取端口转发统计数据 响应结构
type GetPortForwardStatsResponse struct {
	ID            int64     `json:"id"`              // 端口转发统计记录序号
	CreatedAt     time.Time `json:"created_at"`      // 创建时间
	DeviceID      int64     `json:"device_id"`       // 关联 Device 序号
	PortForwardID int64     `json:"port_forward_id"` // 关联 DevicePortForward 序号
	StartTime     time.Time `json:"start_time"`      // 转发启动时间 (UTC)
	ActiveConn    int64     `json:"active_conn"`     // 瞬时活跃连接数
	TotalConn     int64     `json:"total_conn"`      // 累计连接数
	BytesIn       int64     `json:"bytes_in"`        // 累计入站流量 (bytes)
	BytesOut      int64     `json:"bytes_out"`       // 累计出站流量 (bytes)
	UDPPacketsIn  int64     `json:"udp_packets_in"`  // 累计入站 UDP 包数
	UDPPacketsOut int64     `json:"udp_packets_out"` // 累计出站 UDP 包数
	Errors        int64     `json:"errors"`          // 累计错误数
}
