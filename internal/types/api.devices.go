package types

import "time"

// 获取全部设备 响应结构
type GetDeviceResponse struct {
	ID              int64     `json:"id,omitempty"`                // 设备序号
	Code            string    `json:"code"`                        // 设备编号
	Name            string    `json:"name"`                        // 设备名称
	Description     string    `json:"description,omitempty"`       // 设备描述
	IPV4            string    `json:"ip_v4,omitempty"`             // IP地址V4
	IPV6            string    `json:"ip_v6,omitempty"`             // IP地址V6
	LastReportedAt  time.Time `json:"last_reported_at,omitempty"`  // 最后上报时间
	LastHeartbeatAt time.Time `json:"last_heartbeat_at,omitempty"` // 最后心跳时间
}

// 获取设备 请求结构
type GetDeviceRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 设备序号
}

// 创建设备 请求结构
type CreateDeviceRequest struct {
	Code        string `json:"code" binding:"required"` // 设备编号
	Name        string `json:"name" binding:"required"` // 设备名称
	Description string `json:"description,omitempty"`   // 设备描述
	IPV4        string `json:"ip_v4,omitempty"`         // IP地址V4
	IPV6        string `json:"ip_v6,omitempty"`         // IP地址V6
}

// 更新设备 请求结构
type UpdateDeviceRequest struct {
	GetDeviceRequest
	Name        string `json:"name" binding:"required"` // 设备名称
	Description string `json:"description,omitempty"`   // 设备描述
	IPV4        string `json:"ip_v4,omitempty"`         // IP地址V4
	IPV6        string `json:"ip_v6,omitempty"`         // IP地址V6
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
	IPV4      string    `json:"ip_v4"`      // IP地址V4
	IPV6      string    `json:"ip_v6"`      // IP地址V6
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// 检查设备健康状态 请求结构
type HealthcheckDeviceRequest struct {
	GetDeviceRequest
}
