package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// 设备页面 请求结构
type DeviceRequest struct {
	CommonPageRequest
}

// 设备页面 响应结构
type DeviceResponse struct {
	CommonPageResponse[GetDeviceResponse] // 设备列表
}

// 设备端口转发统计 请求结构
type DevicePortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"设备序号"`
}

// 设备端口转发统计 响应结构
type DevicePortForwardStatsResponse struct {
	Device ent.Device               // 关联设备信息
	Rule   ent.DevicePortForward    // 转发规则详情
	Latest ent_log.PortForwardStats // 最后一跳的快照数据
}
