package types

import (
	"vibe_ip/internal/ent"
)

// 同步EdgeOne站点 请求结构
type SyncEdgeOneZonesRequest struct {
	Type     string `json:"type" binding:"required" jsonschema:"同步类型"` // 同步类型
	ConfigID int64  `json:"config_id" jsonschema:"平台配置序号"`             // 平台配置序号
	ZoneID   int64  `json:"zone_id" jsonschema:"站点序号"`                 // 站点序号
	GroupID  int64  `json:"group_id" jsonschema:"源站组序号"`               // 源站组序号
}

// 获取EdgeOne站点 响应结构
type GetEdgeOneZoneResponse struct {
	ID          int64             `json:"id"`          // 平台配置序号
	Capability  string            `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string            `json:"name"`        // 平台配置名称
	Description string            `json:"description"` // 平台配置描述
	Platform    string            `json:"platform"`    // 平台类型
	Zones       []ent.EdgeOneZone `json:"zones"`       // 站点列表
}
