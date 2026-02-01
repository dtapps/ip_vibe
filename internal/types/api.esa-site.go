package types

import (
	"time"
)

// 获取ESA站点 站点
type GetESASiteSite struct {
	ID              int64     `json:"id"`                // ESA 站点序号
	RemoteCreatedAt time.Time `json:"remote_created_at"` // 远程创建时间
	SyncAt          time.Time `json:"sync_at"`           // 最近一次同步时间
	AccessType      string    `json:"access_type"`       // 站点接入类型
	CnameZone       string    `json:"cname_zone"`        // 站点的 CNAME 后缀
	Coverage        string    `json:"coverage"`          // 站点加速区域
	InstanceID      string    `json:"instance_id"`       // 站点绑定的套餐实例 ID
	NameServerList  string    `json:"name_server_list"`  // 站点分配的 NS 列表。使用半角逗号（,）分隔
	PlanName        string    `json:"plan_name"`         // 套餐名称
	PlanSpecName    string    `json:"plan_spec_name"`    // 站点套餐的规格名称
	SiteID          int64     `json:"site_id"`           // 站点 ID
	SiteName        string    `json:"site_name"`         // 站点名称
	Status          string    `json:"status"`            // 站点状态
	OfflineReason   string    `json:"offline_reason"`    // 站点停用原因

	RecordCount     int `json:"record_count"`     // 记录数量
	OriginpoolCount int `json:"originpool_count"` // 源地址池数量
}

// 获取ESA站点 响应结构
type GetESASiteResponse struct {
	ID          int64            `json:"id"`          // 平台配置序号
	Capability  string           `json:"capability"`  // 平台支持的服务能力：all（全部）dns（仅DNS）cdn（仅CDN）ssl（仅SSL）tun（仅TUN）
	Name        string           `json:"name"`        // 平台配置名称
	Description string           `json:"description"` // 平台配置描述
	Platform    string           `json:"platform"`    // 平台类型
	Sites       []GetESASiteSite `json:"sites"`       // 站点列表
}
