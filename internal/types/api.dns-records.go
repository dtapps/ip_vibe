package types

import "time"

// 获取全部DNS记录 请求结构
type GetDnsRecordsRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"title=是否活跃,description=指定是否只返回活跃的DNS记录,example=true"` // 是否活跃
}

// 获取DNS记录 响应结构
type GetDnsRecordResponse struct {
	ID            int64     `json:"id,omitempty" jsonschema:"title=DNS记录序号,description=DNS记录的唯一标识符"`                          // DNS记录序号
	ConfigID      int64     `json:"config_id" jsonschema:"title=平台配置序号,description=关联的平台配置ID"`                                // 平台配置序号
	Platform      string    `json:"platform,omitempty" jsonschema:"title=平台名称,description=平台平台名称"`                            // 平台平台名称
	RecordID      any       `json:"record_id,omitempty" jsonschema:"title=平台侧记录ID,description=平台侧记录的唯一标识符"`                   // 平台侧记录 ID
	Domain        string    `json:"domain" jsonschema:"title=域名,description=域名（如 www.example.com）"`                           // 域名（如 www.example.com）
	Type          string    `json:"type" jsonschema:"title=记录类型,description=记录类型（A、AAAA、CNAME等）,enum=A,enum=AAAA,enum=CNAME"` // 记录类型（A、AAAA、CNAME等）
	IsActive      bool      `json:"is_active" jsonschema:"title=是否活跃,description=是否活跃,default=true"`                          // 是否活跃
	LastRefreshAt time.Time `json:"last_refresh_at,omitempty" jsonschema:"title=最后刷新时间,description=最后一次刷新成功时间"`               // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id" jsonschema:"title=IP组序号,description=关联的IP组序号（与CDN记录序号二选一）"` // IP 组序号（与 cdn_record_id 二选一）
	IPGroupName string `json:"ip_group_name" jsonschema:"title=IP组名称,description=IP组名称"`               // IP 组名称

	CdnRecordID        int64  `json:"cdn_record_id" jsonschema:"title=CDN记录序号,description=关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）"` // CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）
	CdnRecordCdnDomain string `json:"cdn_record_cdn_domain" jsonschema:"title=CDN记录域名,description=CDN记录域名（如 cdn.example.com）"`      // CDN记录域名（如 cdn.example.com）

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取DNS记录 请求结构
type GetDnsRecordRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=DNS记录序号,description=DNS记录的唯一标识符,example=1"` // DNS记录序号
}

// 创建DNS记录 请求结构
type CreateDnsRecordRequest struct {
	ConfigID    int64  `json:"config_id" binding:"required" jsonschema:"title=平台配置序号,description=关联的平台配置ID,example=1"`                                // 平台配置序号
	Domain      string `json:"domain" binding:"required" jsonschema:"title=域名,description=域名（如 www.example.com）,example=www.example.com"`             // 域名（如 www.example.com）
	Type        string `json:"type" binding:"required" jsonschema:"title=记录类型,description=记录类型（A、AAAA、CNAME等）,enum=A,enum=AAAA,enum=CNAME,example=A"` // 记录类型（A、AAAA、CNAME等）
	IPGroupID   int64  `json:"ip_group_id,omitempty" jsonschema:"title=IP组序号,description=关联的IP组序号（与CDN记录序号二选一）,example=1"`                            // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64  `json:"cdn_record_id,omitempty" jsonschema:"title=CDN记录序号,description=关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）,example=1"`      // CDN记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,2,3]"` // 关联的 Webhook 列表
}

// 更新DNS记录 请求结构
type UpdateDnsRecordRequest struct {
	GetDnsRecordRequest
	IPGroupID   int64 `json:"ip_group_id,omitempty" jsonschema:"title=IP组序号,description=关联的IP组序号（与CDN记录序号二选一）,example=2"`                       // IP 组序号（与 cdn_record_id 二选一）
	CdnRecordID int64 `json:"cdn_record_id,omitempty" jsonschema:"title=CDN记录序号,description=关联的CDN记录序号（与IP组序号二选一，一个CDN记录只能被一条DNS引用）,example=2"` // CDN 记录序号（与 ip_group_id 二选一 、 一个 CDN 记录只能被一条 DNS 引用）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,3,4]"` // 关联的 Webhook 列表
}

// 更新DNS记录状态 请求结构
type UpdateDnsRecordActiveRequest struct {
	GetDnsRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"title=是否活跃,description=是否活跃该DNS记录,default=true,example=false"` // 是否活跃
}

// 删除DNS记录 请求结构
type DeleteDnsRecordRequest struct {
	GetDnsRecordRequest
}

// 检查DNS记录健康状态 请求结构
type HealthcheckDnsRecordRequest struct {
	GetDnsRecordRequest
}
