package types

import "vibe_ip/internal/ent"

// 设备组页面 请求结构
type DeviceGroupRequest struct {
	CommonPageRequest
	ID   int64  `json:"id,omitempty" form:"id,omitempty"`     // 设备组序号
	Code string `json:"code,omitempty" form:"code,omitempty"` // 设备组编码
	Name string `json:"name,omitempty" form:"name,omitempty"` // 设备组名称
}

// 设备组页面 响应结构
type DeviceGroupResponse struct {
	CommonPageResponse[GetDeviceGroupResponse, DeviceGroupRequest] // IP组列表
}

// 设备端口转发统计 请求结构
type DeviceGroupPortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"IP组序号"`
}

// 设备端口转发统计 响应结构
type DeviceGroupPortForwardStatsResponse struct {
	Item ent.DeviceGroup // IP组信息
}
