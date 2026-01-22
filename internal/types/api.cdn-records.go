package types

import (
	"encoding/json"
	"time"
)

// 获取全部CDN记录 请求结构
type GetCdnRecordsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 获取CDN记录 响应结构
type GetCdnRecordResponse struct {
	ID            int64           `json:"id" jsonschema:"CDN记录序号"`                        // CDN记录序号
	ConfigID      int64           `json:"config_id" jsonschema:"平台配置序号"`                  // 平台配置序号
	ConfigName    string          `json:"config_name" jsonschema:"平台配置名称"`                // 平台配置名称
	Platform      string          `json:"platform" jsonschema:"平台平台名称"`                   // 平台平台名称
	RecordID      any             `json:"record_id" jsonschema:"平台侧记录ID"`                 // 平台侧记录ID
	DomainName    string          `json:"domain_name" jsonschema:"域名（如 www.example.com）"` // 域名（如 www.example.com）
	SubDomain     string          `json:"sub_domain" jsonschema:"子域名（如 www, blog, api）"`  // 子域名（如 www, blog, api）
	Type          string          `json:"type" jsonschema:"记录类型（A、AAAA）"`                 // 记录类型（A、AAAA）
	ConfigJSON    json.RawMessage `json:"config_json" jsonschema:"CDN记录配置（JSON）"`         // CDN记录配置（JSON）
	CdnDomain     string          `json:"cdn_domain" jsonschema:"CDN边缘域名（平台提供）"`          // CDN 边缘域名（平台提供）
	IsActive      bool            `json:"is_active" jsonschema:"是否活跃"`                    // 是否活跃
	LastRefreshAt time.Time       `json:"last_refresh_at" jsonschema:"最后一次刷新成功时间"`        // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id" jsonschema:"关联的IP组序号"` // IP 组序号
	IPGroupName string `json:"ip_group_name" jsonschema:"IP组名称"`  // IP 组名称

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取CDN记录 请求结构
type GetCdnRecordRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"CDN记录序号"` // CDN记录序号
}

// 创建CDN记录 请求结构
type CreateCdnRecordRequest struct {
	ConfigID   int64  `json:"config_id" binding:"required" jsonschema:"关联的平台配置ID"`              // 平台配置序号
	SubDomain  string `json:"sub_domain" binding:"required" jsonschema:"子域名（如 www, blog, api）"` // 子域名（如 www, blog, api）
	Type       string `json:"type" binding:"required" jsonschema:"记录类型（A、AAAA）"`                // 记录类型（A、AAAA）
	ConfigJSON string `json:"config_json,omitempty" jsonschema:"CDN记录配置（JSON）"`                 // CDN记录配置（JSON）
	IPGroupID  int64  `json:"ip_group_id" binding:"required" jsonschema:"关联的IP组序号"`             // IP 组序号

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新CDN记录 请求结构
type UpdateCdnRecordRequest struct {
	GetCdnRecordRequest
	Type       string `json:"type" binding:"required" jsonschema:"记录类型（A、AAAA）"`    // 记录类型（A、AAAA）
	ConfigJSON string `json:"config_json,omitempty" jsonschema:"CDN记录配置（JSON）"`     // CDN记录配置（JSON）
	IPGroupID  int64  `json:"ip_group_id" binding:"required" jsonschema:"关联的IP组序号"` // IP 组序号

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新CDN记录状态 请求结构
type UpdateCdnRecordActiveRequest struct {
	GetCdnRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 批量更新CDN记录状态 请求结构
type BatchUpdateCdnRecordActiveRequest struct {
	BatchDeleteCdnRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 删除CDN记录 请求结构
type DeleteCdnRecordRequest struct {
	GetCdnRecordRequest
}

// 批量删除CDN记录 请求结构
type BatchDeleteCdnRecordRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"CDN记录的唯一标识符列表"` // CDN记录序号列表
}

// 检查CDN记录健康状态 请求结构
type HealthCheckCdnRecordRequest struct {
	GetCdnRecordRequest
}
