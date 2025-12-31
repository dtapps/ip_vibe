package types

import (
	"vibe_ip/internal/ent"
)

// 设备页面 请求结构
type DeviceRequest struct {
	CommonPageRequest
	ID   int64  `json:"id,omitempty" form:"id,omitempty"`     // 设备序号
	Code string `json:"code,omitempty" form:"code,omitempty"` // 设备编号
	Name string `json:"name,omitempty" form:"name,omitempty"` // 设备名称
}

// 设备页面 响应结构
type DeviceResponse struct {
	CommonPageResponse[GetDeviceResponse, DeviceRequest] // 设备列表
}

// 设备端口转发统计 请求结构
type DevicePortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"设备序号"`
}

// 设备端口转发统计 响应结构
type DevicePortForwardStatsResponse struct {
	Item ent.Device // 设备信息
}
