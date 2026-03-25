package types

import (
	"vibe_ip/internal/ent"
)

// 端口转发页面 请求结构
type PortForwardRequest struct {
	CommonPageRequest
	ID          int64        `json:"id,omitempty" form:"id,omitempty"`           // 设备端口转发序号
	Name        string       `json:"name,omitempty" form:"name,omitempty"`       // 设备端口转发名称
	AutoForward NullableBool `json:"auto_forward" form:"auto_forward,omitempty"` // 防火墙自动放行
	IsActive    NullableBool `json:"is_active" form:"is_active,omitempty"`       // 是否活跃
}

// 端口转发页面 响应结构
type PortForwardResponse struct {
	CommonPageResponse[GetPortForwardResponse, PortForwardRequest] // 端口转发列表
}

// 端口转发统计页面 请求结构
type PortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"端口转发序号"`
}

// 端口转发统计页面 响应结构
type PortForwardStatsResponse struct {
	Item   ent.DevicePortForward // 转发规则详情
	Device ent.Device            // 关联设备信息
}
