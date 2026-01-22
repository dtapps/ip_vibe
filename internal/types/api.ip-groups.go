package types

import "time"

// 获取全部IP组 请求结构
type GetIPGroupsRequest struct {
}

// 获取IP组 响应结构
type GetIPGroupResponse struct {
	ID          int64  `json:"id" jsonschema:"IP组序号"`          // IP组序号
	Code        string `json:"code" jsonschema:"IP组编码"`        // IP组编码
	Name        string `json:"name" jsonschema:"IP组名称"`        // IP组名称
	Description string `json:"description" jsonschema:"IP组描述"` // IP组描述

	DeviceCount int     `json:"device_count" jsonschema:"关联的设备数量"` // 设备数
	Devices     []int64 `json:"devices" jsonschema:"关联的设备ID列表"`    // 设备列表

	CdnRecordCount int     `json:"cdn_record_count" jsonschema:"关联的CDN记录数量"` // CDN记录数
	CdnRecords     []int64 `json:"cdn_records" jsonschema:"关联的CDN记录ID列表"`    // CDN记录列表

	DnsRecordCount int     `json:"dns_record_count" jsonschema:"关联的DNS记录数量"` // DNS记录数
	DnsRecords     []int64 `json:"dns_records" jsonschema:"关联的DNS记录ID列表"`    // DNS记录列表

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取IP组 请求结构
type GetIPGroupRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"IP组的唯一标识符"` // IP组序号
}

// 创建IP组 请求结构
type CreateIPGroupRequest struct {
	Name        string  `json:"name" binding:"required" jsonschema:"IP组的名称"`       // IP组名称
	Description string  `json:"description,omitempty" jsonschema:"IP组的描述信息"`       // IP组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"关联的设备ID列表"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新IP组 请求结构
type UpdateIPGroupRequest struct {
	GetIPGroupRequest
	Name        string  `json:"name" binding:"required" jsonschema:"IP组的名称"`       // IP组名称
	Description string  `json:"description,omitempty" jsonschema:"IP组的描述信息"`       // IP组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"关联的设备ID列表"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 删除IP组 请求结构
type DeleteIPGroupRequest struct {
	GetIPGroupRequest
}

// 批量删除IP组 请求结构
type BatchDeleteIPGroupRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"IP组的唯一标识符列表"` // IP组序号列表
}

// 获取IP组的设备列表 请求结构
type GetIPGroupDevicesRequest struct {
	GetIPGroupRequest
}

// 检查IP组健康状态 请求结构
type HealthCheckIPGroupRequest struct {
	GetIPGroupRequest
}

// 获取IP组设备端口转发统计数据 请求结构
type GetIPGroupDevicePortForwardStatsRequest struct {
	GetIPGroupRequest
	Range string `json:"range" form:"range" binding:"required" jsonschema:"查询时间范围，支持 1h, 6h, 24h, 7d"` // 查询时间范围，支持 1h, 6h, 24h, 7d
}

// 获取IP组设备端口转发统计数据 响应结构
type GetIPGroupDevicePortForwardStatsResponse struct {
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
