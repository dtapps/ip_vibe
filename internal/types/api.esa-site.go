package types

import (
	"vibe_ip/internal/ent"
)

// 获取ESA站点 响应结构
type GetESASiteResponse struct {
	ID          int64         `json:"id"`          // 平台配置序号
	Capability  string        `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string        `json:"name"`        // 平台配置名称
	Description string        `json:"description"` // 平台配置描述
	Platform    string        `json:"platform"`    // 平台类型
	Sites       []ent.EsaSite `json:"sites"`       // 站点列表
}
