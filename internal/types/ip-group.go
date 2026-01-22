package types

import "vibe_ip/internal/ent"

// IP组页面 请求结构
type IPGroupRequest struct {
	CommonPageRequest
	ID   int64  `json:"id,omitempty" form:"id,omitempty"`     // IP组序号
	Code string `json:"code,omitempty" form:"code,omitempty"` // IP组编码
	Name string `json:"name,omitempty" form:"name,omitempty"` // IP组名称
}

// IP组页面 响应结构
type IPGroupResponse struct {
	CommonPageResponse[GetIPGroupResponse, IPGroupRequest] // IP组列表
}

// 设备端口转发统计 请求结构
type IPGroupPortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"IP组序号"`
}

// 设备端口转发统计 响应结构
type IPGroupPortForwardStatsResponse struct {
	Item ent.IpGroup // IP组信息
}
