package types

import "time"

// 获取全部设备 请求结构
type GetDevicesRequest struct {
}

// 获取设备 响应结构
type GetDeviceResponse struct {
	ID              int64     `json:"id,omitempty" jsonschema:"设备的唯一标识符"`      // 设备序号
	Code            string    `json:"code" jsonschema:"设备的编号"`                 // 设备编号
	Name            string    `json:"name" jsonschema:"设备的名称"`                 // 设备名称
	Description     string    `json:"description" jsonschema:"设备的描述信息"`        // 设备描述
	IPV4            string    `json:"ip_v4" jsonschema:"设备的IPv4地址"`            // IP地址V4
	IPV6            string    `json:"ip_v6" jsonschema:"设备的IPv6地址"`            // IP地址V6
	LastReportedAt  time.Time `json:"last_reported_at" jsonschema:"设备最后上报时间"`  // 最后上报时间
	LastHeartbeatAt time.Time `json:"last_heartbeat_at" jsonschema:"设备最后心跳时间"` // 最后心跳时间

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表

	Online bool   `json:"online" jsonschema:"设备是否在线"` // 设备是否在线
	Status string `json:"status" jsonschema:"设备的状态"`  // 设备状态

	CreatedAt time.Time `json:"-"` // 创建时间
	UpdatedAt time.Time `json:"-"` // 更新时间
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

// 批量删除设备 请求结构
type BatchDeleteDeviceRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"设备的唯一标识符列表"` // 设备序号列表
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

// 获取设备所有端口转发统计数据 最新状态快照
type GetDevicePortForwardStatsLatest struct {
	ActiveConn int64 `json:"active_conn"` // 当前活跃连接数
	BytesIn    int64 `json:"bytes_in"`    // 累计入站流量 (bytes)
	BytesOut   int64 `json:"bytes_out"`   // 累计出站流量 (bytes)
	Errors     int64 `json:"errors"`      // 累计错误数
}

// 获取设备所有端口转发统计数据 时间范围内的增量统计
type GetDevicePortForwardStatsDelta struct {
	BytesInIncr  int64 `json:"bytes_in_incr"`  // 本时段新增入站流量 (bytes)
	BytesOutIncr int64 `json:"bytes_out_incr"` // 本时段新增出站流量 (bytes)
	ErrorsIncr   int64 `json:"errors_incr"`    // 本时段新增错误数
}

// 获取设备所有端口转发统计数据 单个时间点的速率数据
type GetDevicePortForwardStatsRates struct {
	Timestamp  int64   `json:"timestamp"`   // 数据采样时间戳（UTC）
	InKBps     float64 `json:"in_kbps"`     // 入站速率（KB/s，保留3位小数）
	OutKBps    float64 `json:"out_kbps"`    // 出站速率（KB/s，保留3位小数）
	ActiveConn int64   `json:"active_conn"` // 该时刻的活跃连接数
}

// 获取设备所有端口转发统计数据 响应结构
type GetDevicePortForwardStatsResponse struct {
	PortIDs []int64                          `json:"port_ids"` // 关联端口ID列表
	Latest  GetDevicePortForwardStatsLatest  `json:"latest"`   // 最新状态快照
	Delta   GetDevicePortForwardStatsDelta   `json:"delta"`    // 所选时间范围内的增量
	Rates   []GetDevicePortForwardStatsRates `json:"rates"`    // 速率趋势数据
}

// 获取设备状态 请求结构
type GetDeviceStatusRequest struct {
	GetDeviceRequest
}

// 获取设备状态 响应结构
type GetDeviceStatusResponse struct {
	Role      string `json:"role" jsonschema:"角色"`            // 角色
	Version   string `json:"version" jsonschema:"版本号"`        // 版本号
	GitCommit string `json:"git_commit" jsonschema:"git提交哈希"` // git提交哈希
	BuildTime string `json:"build_time" jsonschema:"构建时间"`    // 构建时间
	GoVersion string `json:"go_version" jsonschema:"go版本"`    // go版本
	Os        string `json:"os" jsonschema:"操作系统"`            // 操作系统
	Arch      string `json:"arch" jsonschema:"架构"`            // 架构
}
