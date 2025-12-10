package types

import "time"

// 获取设备 响应结构
type GetDeviceResponse struct {
	ID              int64     `json:"id,omitempty" jsonschema:"title=设备序号,description=设备的唯一标识符"`                  // 设备序号
	Code            string    `json:"code" jsonschema:"title=设备编号,description=设备的编号"`                             // 设备编号
	Name            string    `json:"name" jsonschema:"title=设备名称,description=设备的名称"`                             // 设备名称
	Description     string    `json:"description,omitempty" jsonschema:"title=设备描述,description=设备的描述信息"`          // 设备描述
	IPV4            string    `json:"ip_v4,omitempty" jsonschema:"title=IPv4地址,description=设备的IPv4地址"`            // IP地址V4
	IPV6            string    `json:"ip_v6,omitempty" jsonschema:"title=IPv6地址,description=设备的IPv6地址"`            // IP地址V6
	LastReportedAt  time.Time `json:"last_reported_at,omitempty" jsonschema:"title=最后上报时间,description=设备最后上报时间"`  // 最后上报时间
	LastHeartbeatAt time.Time `json:"last_heartbeat_at,omitempty" jsonschema:"title=最后心跳时间,description=设备最后心跳时间"` // 最后心跳时间

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取设备 请求结构
type GetDeviceRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=设备序号,description=设备的唯一标识符,example=1"` // 设备序号
}

// 创建设备 请求结构
type CreateDeviceRequest struct {
	Code        string `json:"code" binding:"required" jsonschema:"title=设备编号,description=设备的编号,example=device-001"`                            // 设备编号
	Name        string `json:"name" binding:"required" jsonschema:"title=设备名称,description=设备的名称,example=服务器001"`                                // 设备名称
	Description string `json:"description,omitempty" jsonschema:"title=设备描述,description=设备的描述信息,example=用于监控网络流量"`                              // 设备描述
	IPV4        string `json:"ip_v4,omitempty" jsonschema:"title=IPv4地址,description=设备的IPv4地址,example=192.168.1.100"`                           // IP地址V4
	IPV6        string `json:"ip_v6,omitempty" jsonschema:"title=IPv6地址,description=设备的IPv6地址,example=2001:0db8:85a3:0000:0000:8a2e:0370:7334"` // IP地址V6

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,2,3]"` // 关联的 Webhook 列表
}

// 更新设备 请求结构
type UpdateDeviceRequest struct {
	GetDeviceRequest
	Name        string `json:"name" binding:"required" jsonschema:"title=设备名称,description=设备的名称,example=更新后的服务器001"`                            // 设备名称
	Description string `json:"description,omitempty" jsonschema:"title=设备描述,description=设备的描述信息,example=更新后的设备描述"`                              // 设备描述
	IPV4        string `json:"ip_v4,omitempty" jsonschema:"title=IPv4地址,description=设备的IPv4地址,example=192.168.1.101"`                           // IP地址V4
	IPV6        string `json:"ip_v6,omitempty" jsonschema:"title=IPv6地址,description=设备的IPv6地址,example=2001:0db8:85a3:0000:0000:8a2e:0370:7335"` // IP地址V6

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,3,4]"` // 关联的 Webhook 列表
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
	IPV4      string    `json:"ip_v4" jsonschema:"title=IPv4地址,description=设备的IPv4地址"`  // IP地址V4
	IPV6      string    `json:"ip_v6" jsonschema:"title=IPv6地址,description=设备的IPv6地址"`  // IP地址V6
	CreatedAt time.Time `json:"created_at" jsonschema:"title=创建时间,description=记录创建的时间"` // 创建时间
}

// 检查设备健康状态 请求结构
type HealthcheckDeviceRequest struct {
	GetDeviceRequest
}
