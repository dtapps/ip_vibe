package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// 端口转发页面 请求结构
type PortForwardRequest struct {
	CommonPageRequest
}

// 端口转发页面 响应结构
type PortForwardResponse struct {
	CommonPageResponse[GetPortForwardResponse] // 端口转发列表
}

// 端口转发统计页面 请求结构
type PortForwardStatsRequest struct {
	ID int64 `json:"id" form:"id" validate:"required" comment:"端口转发序号"`
}

// 端口转发统计页面 响应结构
type PortForwardStatsResponse struct {
	Item        ent.DevicePortForward    // 转发规则详情
	Device      ent.Device               // 关联设备信息
	LatestStats ent_log.PortForwardStats // 最后一跳的快照数据
}
