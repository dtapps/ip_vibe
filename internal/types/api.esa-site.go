package types

import (
	"vibe_ip/internal/ent"
)

// 同步ESA站点 请求结构
type SyncESASitesRequest struct {
	Type     string `json:"type" binding:"required" jsonschema:"同步类型"` // 同步类型
	ConfigID int64  `json:"config_id" jsonschema:"平台配置序号"`             // 平台配置序号
	SiteID   int64  `json:"site_id" jsonschema:"站点序号"`                 // 站点序号
	PoolID   int64  `json:"pool_id" jsonschema:"源地址池序号"`               // 源地址池序号
}

// 获取ESA站点 响应结构
type GetESASiteResponse struct {
	ID          int64         `json:"id"`          // 平台配置序号
	Capability  string        `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string        `json:"name"`        // 平台配置名称
	Description string        `json:"description"` // 平台配置描述
	Platform    string        `json:"platform"`    // 平台类型
	Sites       []ent.EsaSite `json:"sites"`       // 站点列表
}
