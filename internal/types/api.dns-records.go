package types

import (
	"encoding/json"
	"time"
)

// 获取全部DNS记录 请求结构
type GetDnsRecordsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 获取DNS记录 响应结构
type GetDnsRecordResponse struct {
	ID            int64           `json:"id" jsonschema:"DNS记录序号"`                          // DNS记录序号
	ConfigID      int64           `json:"config_id" jsonschema:"关联的平台配置ID"`                 // 平台配置序号
	ConfigName    string          `json:"config_name" jsonschema:"平台配置序号"`                  // 平台配置名称
	Platform      string          `json:"platform" jsonschema:"平台平台名称"`                     // 平台平台名称
	RecordID      any             `json:"record_id,omitempty" jsonschema:"平台侧记录ID"`         // 平台侧记录ID
	DomainName    string          `json:"domain_name" jsonschema:"域名（如 www.example.com）"`   // 域名（如 www.example.com）
	SubDomain     string          `json:"sub_domain" jsonschema:"子域名（如 www, blog, api）"`    // 子域名（如 www, blog, api）
	Type          string          `json:"type" jsonschema:"记录类型（A、AAAA、CNAME）"`             // 记录类型（A、AAAA、CNAME）
	ConfigJSON    json.RawMessage `json:"config_json,omitempty" jsonschema:"DNS记录配置（JSON）"` // DNS记录配置（JSON）
	IsActive      bool            `json:"is_active" jsonschema:"是否活跃"`                      // 是否活跃
	LastRefreshAt time.Time       `json:"last_refresh_at" jsonschema:"最后一次刷新成功时间"`          // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id" jsonschema:"关联的IP组序号（与CDN记录序号二选一）"` // IP 组序号（与 cdn_record_id 二选一）
	IPGroupName string `json:"ip_group_name" jsonschema:"IP组名称"`               // IP 组名称

	CdnRecordID        int64  `json:"cdn_record_id" jsonschema:"关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）"` // CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）
	CdnRecordCdnDomain string `json:"cdn_record_cdn_domain" jsonschema:"CDN记录域名（如 cdn.example.com）"`      // CDN记录域名（如 cdn.example.com）

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取DNS记录 请求结构
type GetDnsRecordRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"DNS记录序号"` // DNS记录序号
}

// 创建DNS记录 请求结构
type CreateDnsRecordRequest struct {
	ConfigID    int64  `json:"config_id" binding:"required" jsonschema:"关联的平台配置ID"`                          // 平台配置序号
	SubDomain   string `json:"sub_domain" binding:"required" jsonschema:"子域名（如 www, blog, api）"`             // 子域名（如 www, blog, api）
	Type        string `json:"type" binding:"required" jsonschema:"记录类型（A、AAAA、CNAME）"`                      // 记录类型（A、AAAA、CNAME）
	ConfigJSON  string `json:"config_json,omitempty" jsonschema:"DNS记录配置（JSON）"`                             // DNS记录配置（JSON）
	IPGroupID   int64  `json:"ip_group_id,omitempty" jsonschema:"关联的IP组序号（与CDN记录序号二选一）"`                     // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64  `json:"cdn_record_id,omitempty" jsonschema:"关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）"` // CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新DNS记录 请求结构
type UpdateDnsRecordRequest struct {
	GetDnsRecordRequest
	Type        string `json:"type" binding:"required" jsonschema:"记录类型（A、AAAA、CNAME）"`                      // 记录类型（A、AAAA、CNAME）
	ConfigJSON  string `json:"config_json,omitempty" jsonschema:"DNS记录配置（JSON）"`                             // DNS记录配置（JSON）
	IPGroupID   int64  `json:"ip_group_id,omitempty" jsonschema:"关联的IP组序号（与CDN记录序号二选一）"`                     // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64  `json:"cdn_record_id,omitempty" jsonschema:"关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）"` // CDN 记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新DNS记录状态 请求结构
type UpdateDnsRecordActiveRequest struct {
	GetDnsRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃该DNS记录"` // 是否活跃
}

// 批量更新DNS记录状态 请求结构
type BatchUpdateDnsRecordActiveRequest struct {
	BatchDeleteDnsRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 删除DNS记录 请求结构
type DeleteDnsRecordRequest struct {
	GetDnsRecordRequest
}

// 批量删除DNS记录 请求结构
type BatchDeleteDnsRecordRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"DNS记录的唯一标识符列表"` // DNS记录序号列表
}

// 检查DNS记录健康状态 请求结构
type HealthCheckDnsRecordRequest struct {
	GetDnsRecordRequest
}
