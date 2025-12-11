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
