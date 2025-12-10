package types

import (
	"time"
)

// 获取全部CDN记录 请求结构
type GetCdnRecordsRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"title=是否活跃,description=指定是否只返回活跃的CDN记录,example=true"` // 是否活跃
}

// 获取CDN记录 响应结构
type GetCdnRecordResponse struct {
	ID             int64     `json:"id,omitempty" jsonschema:"title=CDN记录序号,description=CDN记录的唯一标识符"`                                              // CDN记录序号
	ConfigID       int64     `json:"config_id" jsonschema:"title=平台配置序号,description=关联的平台配置ID"`                                                    // 平台配置序号
	Platform       string    `json:"platform,omitempty" jsonschema:"title=平台名称,description=平台平台名称"`                                                // 平台平台名称
	RecordID       any       `json:"record_id,omitempty" jsonschema:"title=平台侧记录ID,description=平台侧记录的唯一标识符"`                                       // 平台侧记录 ID
	Domain         string    `json:"domain" jsonschema:"title=域名,description=域名（如 www.example.com）"`                                               // 域名（如 www.example.com）
	Type           string    `json:"type" jsonschema:"title=记录类型,description=记录类型（A、AAAA等）,enum=A,enum=AAAA"`                                      // 记录类型（A、AAAA等）
	OriginProtocol string    `json:"origin_protocol" jsonschema:"title=回源协议,description=回源协议（http、https、follow）,enum=http,enum=https,enum=follow"` // 回源协议（http、https、follow）
	HTTPPort       int       `json:"http_port" jsonschema:"title=HTTP回源端口,description=HTTP回源端口,default=80"`                                        // HTTP 回源端口
	HTTPSPort      int       `json:"https_port" jsonschema:"title=HTTPS回源端口,description=HTTPS回源端口,default=443"`                                    // HTTPS 回源端口
	CdnDomain      string    `json:"cdn_domain,omitempty" jsonschema:"title=CDN边缘域名,description=CDN边缘域名（平台提供）"`                                    // CDN 边缘域名（平台提供）
	IsActive       bool      `json:"is_active" jsonschema:"title=是否活跃,description=是否活跃,default=true"`                                              // 是否活跃
	LastRefreshAt  time.Time `json:"last_refresh_at,omitempty" jsonschema:"title=最后刷新时间,description=最后一次刷新成功时间"`                                   // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id" jsonschema:"title=IP组序号,description=关联的IP组序号"` // IP 组序号
	IPGroupName string `json:"ip_group_name" jsonschema:"title=IP组名称,description=IP组名称"`  // IP 组名称

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取CDN记录 请求结构
type GetCdnRecordRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required" jsonschema:"title=CDN记录序号,description=CDN记录的唯一标识符,example=1"` // CDN记录序号
}

// 创建CDN记录 请求结构
type CreateCdnRecordRequest struct {
	ConfigID       int64  `json:"config_id" binding:"required" jsonschema:"title=平台配置序号,description=关联的平台配置ID,example=1"`                                                        // 平台配置序号
	Domain         string `json:"domain" binding:"required" jsonschema:"title=域名,description=域名（如 www.example.com）,example=www.example.com"`                                     // 域名（如 www.example.com）
	Type           string `json:"type" binding:"required" jsonschema:"title=记录类型,description=记录类型（A、AAAA等）,enum=A,enum=AAAA,example=A"`                                          // 记录类型（A、AAAA等）
	OriginProtocol string `json:"origin_protocol" binding:"required" jsonschema:"title=回源协议,description=回源协议（http、https、follow）,enum=http,enum=https,enum=follow,example=https"` // 回源协议（http、https、follow）
	HTTPPort       int    `json:"http_port" binding:"required" jsonschema:"title=HTTP回源端口,description=HTTP回源端口,default=80,example=80"`                                           // HTTP 回源端口
	HTTPSPort      int    `json:"https_port" binding:"required" jsonschema:"title=HTTPS回源端口,description=HTTPS回源端口,default=443,example=443"`                                      // HTTPS 回源端口
	IPGroupID      int64  `json:"ip_group_id" binding:"required" jsonschema:"title=IP组序号,description=关联的IP组序号,example=1"`                                                        // IP 组序号

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,2,3]"` // 关联的 Webhook 列表
}

// 更新CDN记录 请求结构
type UpdateCdnRecordRequest struct {
	GetCdnRecordRequest
	OriginProtocol string `json:"origin_protocol" binding:"required" jsonschema:"title=回源协议,description=回源协议（http、https、follow）,enum=http,enum=https,enum=follow,example=https"` // 回源协议（http、https、follow）
	HTTPPort       int    `json:"http_port" binding:"required" jsonschema:"title=HTTP回源端口,description=HTTP回源端口,default=80,example=80"`                                           // HTTP 回源端口
	HTTPSPort      int    `json:"https_port" binding:"required" jsonschema:"title=HTTPS回源端口,description=HTTPS回源端口,default=443,example=443"`                                      // HTTPS 回源端口
	IPGroupID      int64  `json:"ip_group_id" binding:"required" jsonschema:"title=IP组序号,description=关联的IP组序号,example=1"`                                                        // IP 组序号

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"title=关联的Webhook列表,description=关联的Webhook ID列表,example=[1,3,4]"` // 关联的 Webhook 列表
}

// 更新CDN记录状态 请求结构
type UpdateCdnRecordActiveRequest struct {
	GetCdnRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"title=是否活跃,description=是否活跃,default=true,example=true"` // 是否活跃
}

// 删除CDN记录 请求结构
type DeleteCdnRecordRequest struct {
	GetCdnRecordRequest
}

// 检查CDN记录健康状态 请求结构
type HealthcheckCdnRecordRequest struct {
	GetCdnRecordRequest
}
