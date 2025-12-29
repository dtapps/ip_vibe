package types

import "time"

// 获取全部设备 请求结构
type GetDevicesRequest struct {
}

// 获取设备 响应结构
type GetDeviceResponse struct {
	ID              int64     `json:"id,omitempty" jsonschema:"设备的唯一标识符"`                // 设备序号
	Code            string    `json:"code" jsonschema:"设备的编号"`                           // 设备编号
	Name            string    `json:"name" jsonschema:"设备的名称"`                           // 设备名称
	Description     string    `json:"description,omitempty" jsonschema:"设备的描述信息"`        // 设备描述
	IPV4            string    `json:"ip_v4,omitempty" jsonschema:"设备的IPv4地址"`            // IP地址V4
	IPV6            string    `json:"ip_v6,omitempty" jsonschema:"设备的IPv6地址"`            // IP地址V6
	LastReportedAt  time.Time `json:"last_reported_at,omitempty" jsonschema:"设备最后上报时间"`  // 最后上报时间
	LastHeartbeatAt time.Time `json:"last_heartbeat_at,omitempty" jsonschema:"设备最后心跳时间"` // 最后心跳时间

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取设备 请求结构
type GetDeviceRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"设备的唯一标识符"` // 设备序号
}

// 创建设备 请求结构
type CreateDeviceRequest struct {
	Code        string `json:"code" binding:"required" jsonschema:"设备的编号"` // 设备编号
	Name        string `json:"name" binding:"required" jsonschema:"设备的名称"` // 设备名称
	Description string `json:"description,omitempty" jsonschema:"设备的描述信息"` // 设备描述
	IPV4        string `json:"ip_v4,omitempty" jsonschema:"设备的IPv4地址"`     // IP地址V4
	IPV6        string `json:"ip_v6,omitempty" jsonschema:"设备的IPv6地址"`     // IP地址V6

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新设备 请求结构
type UpdateDeviceRequest struct {
	GetDeviceRequest
	Name        string `json:"name" binding:"required" jsonschema:"设备的名称"` // 设备名称
	Description string `json:"description,omitempty" jsonschema:"设备的描述信息"` // 设备描述
	IPV4        string `json:"ip_v4,omitempty" jsonschema:"设备的IPv4地址"`     // IP地址V4
	IPV6        string `json:"ip_v6,omitempty" jsonschema:"设备的IPv6地址"`     // IP地址V6

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 删除设备 请求结构
type DeleteDeviceRequest struct {
	GetDeviceRequest
}

// 删除设备 获取设备的IP历史记录
type GetDeviceIpHistoryRequest struct {
	GetDeviceRequest
}

// 获取设备的IP历史记录 响应结构
type GetDeviceIpHistoryResponse struct {
	IPV4      string    `json:"ip_v4" jsonschema:"设备的IPv4地址"`    // IP地址V4
	IPV6      string    `json:"ip_v6" jsonschema:"设备的IPv6地址"`    // IP地址V6
	CreatedAt time.Time `json:"created_at" jsonschema:"记录创建的时间"` // 创建时间
}

// 检查设备健康状态 请求结构
type HealthCheckDeviceRequest struct {
	GetDeviceRequest
}

// 获取设备端口转发统计数据 请求结构
type GetDevicePortForwardStatsRequest struct {
	GetDeviceRequest
	Range string `json:"range" form:"range" binding:"required" jsonschema:"查询时间范围，支持 1h, 6h, 24h, 7d"` // 查询时间范围，支持 1h, 6h, 24h, 7d
}

// 获取设备端口转发统计数据 响应结构
type GetDevicePortForwardStatsResponse struct {
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
