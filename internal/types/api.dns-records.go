package types

import "time"

// 获取DNS记录 响应结构
type GetDnsRecordResponse struct {
	ID            int64     `json:"id,omitempty"`              // DNS 记录序号
	ConfigID      int64     `json:"config_id"`                 // 平台配置序号
	Platform      string    `json:"platform,omitempty"`        // 平台平台名称
	RecordID      any       `json:"record_id,omitempty"`       // 平台侧记录 ID
	Domain        string    `json:"domain"`                    // 域名（如 www.example.com）
	Type          string    `json:"type"`                      // 记录类型（A、AAAA、CNAME等）
	IsActive      bool      `json:"is_active"`                 // 是否活跃
	LastRefreshAt time.Time `json:"last_refresh_at,omitempty"` // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id"`   // IP 组序号（与 cdn_record_id 二选一）
	IPGroupName string `json:"ip_group_name"` // IP 组名称

	CdnRecordID        int64  `json:"cdn_record_id"`         // CDN 记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）
	CdnRecordCdnDomain string `json:"cdn_record_cdn_domain"` // CDN 记录域名（如 cdn.example.com）

	Webhooks []int64 `json:"webhooks,omitempty"` // 关联的 Webhook 列表
}

// 获取DNS记录 请求结构
type GetDnsRecordRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // DNS 记录序号
}

// 创建DNS记录 请求结构
type CreateDnsRecordRequest struct {
	ConfigID    int64  `json:"config_id" binding:"required"` // 平台配置序号
	Domain      string `json:"domain" binding:"required"`    // 域名（如 www.example.com）
	Type        string `json:"type" binding:"required"`      // 记录类型（A、AAAA、CNAME等）
	IPGroupID   int64  `json:"ip_group_id,omitempty"`        // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64  `json:"cdn_record_id,omitempty"`      // CDN 记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty"` // 关联的 Webhook 列表
}

// 更新DNS记录 请求结构
type UpdateDnsRecordRequest struct {
	GetDnsRecordRequest
	IPGroupID   int64 `json:"ip_group_id,omitempty"`   // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64 `json:"cdn_record_id,omitempty"` // CDN 记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty"` // 关联的 Webhook 列表
}

// 更新DNS记录状态 请求结构
type UpdateDnsRecordActiveRequest struct {
	GetDnsRecordRequest
	IsActive bool `json:"is_active,omitempty"` // 是否活跃
}

// 删除DNS记录 请求结构
type DeleteDnsRecordRequest struct {
	GetDnsRecordRequest
}

// 检查DNS记录健康状态 请求结构
type HealthcheckDnsRecordRequest struct {
	GetDnsRecordRequest
}
