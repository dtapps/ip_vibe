package types

import (
	"time"
)

// 获取EdgeOne 站点
type GetEdgeOneZone struct {
	ID              int64     `json:"id"`                // EdgeOne 站点序号
	RemoteCreatedAt time.Time `json:"remote_created_at"` // 远程创建时间
	SyncAt          time.Time `json:"sync_at"`           // 最近一次同步时间
	ZoneID          string    `json:"zone_id"`           // 站点ID
	ZoneName        string    `json:"zone_name"`         // 站点名称
	Area            string    `json:"area"`              // 加速区域
	Type            string    `json:"type"`              // 接入类型
	Status          string    `json:"status"`            // 状态
	CnameStatus     string    `json:"cname_status"`      // CNAME 接入状态
	ActiveStatus    string    `json:"active_status"`     // 展示状态
	Paused          bool      `json:"paused"`            // 是否关闭

	AccelerationdomainCount int `json:"accelerationdomain_count"` // 加速域名数量
	DnsrecordCount          int `json:"dnsrecord_count"`          // DNS记录数量
	OrigingroupCount        int `json:"origingroup_count"`        // 源站组数量
}

// 获取EdgeOne 响应结构
type GetEdgeOneResponse struct {
	ID          int64            `json:"id"`          // 平台配置序号
	Capability  string           `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string           `json:"name"`        // 平台配置名称
	Description string           `json:"description"` // 平台配置描述
	Platform    string           `json:"platform"`    // 平台类型
	Zones       []GetEdgeOneZone `json:"zones"`       // 站点列表
}
