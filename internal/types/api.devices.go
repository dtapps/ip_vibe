package types

import "time"

// 获取全部设备 响应结构
type GetDevicesResponse struct {
	ID              int64     `json:"id,omitempty"`                // 设备序号
	Code            string    `json:"code"`                        // 设备编号
	Name            string    `json:"name"`                        // 设备名称
	Description     string    `json:"description,omitempty"`       // 设备描述
	IPV4            string    `json:"ip_v4,omitempty"`             // IP地址V4
	IPV6            string    `json:"ip_v6,omitempty"`             // IP地址V6
	LastReportedAt  time.Time `json:"last_reported_at,omitempty"`  // 最后上报时间
	LastHeartbeatAt time.Time `json:"last_heartbeat_at,omitempty"` // 最后心跳时间
}

// 创建设备 请求结构
type CreateDeviceRequest struct {
	Code        string `json:"code"`                  // 设备编号
	Name        string `json:"name"`                  // 设备名称
	Description string `json:"description,omitempty"` // 设备描述
}

// 更新设备 请求结构
type UpdateDeviceRequest struct {
	Name        string `json:"name"`                  // 设备名称
	Description string `json:"description,omitempty"` // 设备描述
}

// 获取设备的IP历史记录 响应结构
type GetDeviceIpHistoryRequest struct {
	IPV4      string    `json:"ip_v4"`      // IP地址V4
	IPV6      string    `json:"ip_v6"`      // IP地址V6
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
