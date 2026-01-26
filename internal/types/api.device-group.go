package types

import "time"

// 获取全部设备组 请求结构
type GetDeviceGroupsRequest struct {
}

// 获取设备组 响应结构
type GetDeviceGroupResponse struct {
	ID          int64  `json:"id" jsonschema:"设备组序号"`          // 设备组序号
	Code        string `json:"code" jsonschema:"设备组编码"`        // 设备组编码
	Name        string `json:"name" jsonschema:"设备组名称"`        // 设备组名称
	Description string `json:"description" jsonschema:"设备组描述"` // 设备组描述

	DeviceCount int     `json:"device_count" jsonschema:"关联的设备数量"` // 设备数
	Devices     []int64 `json:"devices" jsonschema:"关联的设备ID列表"`    // 设备列表

	CdnRecordCount int     `json:"cdn_record_count" jsonschema:"关联的CDN记录数量"` // CDN记录数
	CdnRecords     []int64 `json:"cdn_records" jsonschema:"关联的CDN记录ID列表"`    // CDN记录列表

	DnsRecordCount int     `json:"dns_record_count" jsonschema:"关联的DNS记录数量"` // DNS记录数
	DnsRecords     []int64 `json:"dns_records" jsonschema:"关联的DNS记录ID列表"`    // DNS记录列表

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表

	CreatedAt time.Time `json:"-"` // 创建时间
	UpdatedAt time.Time `json:"-"` // 更新时间
}

// 获取设备组 请求结构
type GetDeviceGroupRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"设备组的唯一标识符"` // 设备组序号
}

// 创建设备组 请求结构
type CreateDeviceGroupRequest struct {
	Name        string  `json:"name" binding:"required" jsonschema:"设备组的名称"`       // 设备组名称
	Description string  `json:"description,omitempty" jsonschema:"设备组的描述信息"`       // 设备组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"关联的设备ID列表"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新设备组 请求结构
type UpdateDeviceGroupRequest struct {
	GetDeviceGroupRequest
	Name        string  `json:"name" binding:"required" jsonschema:"设备组的名称"`       // 设备组名称
	Description string  `json:"description,omitempty" jsonschema:"设备组的描述信息"`       // 设备组描述
	Devices     []int64 `json:"devices" binding:"required" jsonschema:"关联的设备ID列表"` // 设备列表

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 删除设备组 请求结构
type DeleteDeviceGroupRequest struct {
	GetDeviceGroupRequest
}

// 批量删除设备组 请求结构
type BatchDeleteDeviceGroupRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"设备组的唯一标识符列表"` // 设备组序号列表
}

// 获取设备组的设备列表 请求结构
type GetDeviceGroupDevicesRequest struct {
	GetDeviceGroupRequest
}

// 检查设备组健康状态 请求结构
type HealthCheckDeviceGroupRequest struct {
	GetDeviceGroupRequest
}

// 获取设备组设备端口转发统计数据 请求结构
type GetDeviceGroupDevicePortForwardStatsRequest struct {
	GetDeviceGroupRequest
	Range string `json:"range" form:"range" binding:"required" jsonschema:"查询时间范围，支持 1h, 6h, 24h, 7d"` // 查询时间范围，支持 1h, 6h, 24h, 7d
}

// 获取设备所有端口转发统计数据 最新状态快照
type GetDeviceGroupDevicePortForwardStatsLatest struct {
	ActiveConn int64 `json:"active_conn"` // 当前活跃连接数
	BytesIn    int64 `json:"bytes_in"`    // 累计入站流量 (bytes)
	BytesOut   int64 `json:"bytes_out"`   // 累计出站流量 (bytes)
	Errors     int64 `json:"errors"`      // 累计错误数
}

// 获取设备所有端口转发统计数据 时间范围内的增量统计
type GetDeviceGroupDevicePortForwardStatsDelta struct {
	BytesInIncr  int64 `json:"bytes_in_incr"`  // 本时段新增入站流量 (bytes)
	BytesOutIncr int64 `json:"bytes_out_incr"` // 本时段新增出站流量 (bytes)
	ErrorsIncr   int64 `json:"errors_incr"`    // 本时段新增错误数
}

// 获取设备所有端口转发统计数据 单个时间点的速率数据
type GetDeviceGroupDevicePortForwardStatsRates struct {
	Timestamp  int64   `json:"timestamp"`   // 数据采样时间戳（UTC）
	InKBps     float64 `json:"in_kbps"`     // 入站速率（KB/s，保留3位小数）
	OutKBps    float64 `json:"out_kbps"`    // 出站速率（KB/s，保留3位小数）
	ActiveConn int64   `json:"active_conn"` // 该时刻的活跃连接数
}

// 获取设备组设备端口转发统计数据 响应结构
type GetDeviceGroupDevicePortForwardStatsResponse struct {
	DeviceIDs []int64                                     `json:"device_ids"` // 关联的设备ID列表
	Latest    GetDeviceGroupDevicePortForwardStatsLatest  `json:"latest"`     // 最新状态快照
	Delta     GetDeviceGroupDevicePortForwardStatsDelta   `json:"delta"`      // 所选时间范围内的增量
	Rates     []GetDeviceGroupDevicePortForwardStatsRates `json:"rates"`      // 速率趋势数据
}
