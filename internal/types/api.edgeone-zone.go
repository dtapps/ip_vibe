package types

import (
	"vibe_ip/internal/ent"
)

// 获取EdgeOne站点 响应结构
type GetEdgeOneZoneResponse struct {
	ID          int64             `json:"id"`          // 平台配置序号
	Capability  string            `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string            `json:"name"`        // 平台配置名称
	Description string            `json:"description"` // 平台配置描述
	Platform    string            `json:"platform"`    // 平台类型
	Zones       []ent.EdgeOneZone `json:"zones"`       // 站点列表
}
