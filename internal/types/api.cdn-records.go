package types

import "time"

// 获取全部CDN记录 响应结构
type GetCdnRecordResponse struct {
	ID             int64     `json:"id,omitempty"`              // CDN 记录序号
	ConfigID       int64     `json:"config_id"`                 // 平台配置序号
	Platform       string    `json:"platform,omitempty"`        // 平台平台名称
	RecordID       any       `json:"record_id,omitempty"`       // 平台侧记录 ID
	Domain         string    `json:"domain"`                    // 域名（如 www.example.com）
	Type           string    `json:"type"`                      // 记录类型（A、AAAA等）
	OriginProtocol string    `json:"origin_protocol"`           // 回源协议（http、https、follow）
	HTTPPort       int       `json:"http_port"`                 // HTTP 回源端口
	HTTPSPort      int       `json:"https_port"`                // HTTPS 回源端口
	CdnDomain      string    `json:"cdn_domain,omitempty"`      // CDN 边缘域名（平台提供）
	IsActive       bool      `json:"is_active"`                 // 是否活跃
	LastRefreshAt  time.Time `json:"last_refresh_at,omitempty"` // 最后一次刷新成功时间

	IPGroupID   int64  `json:"ip_group_id"`   // IP 组序号
	IPGroupName string `json:"ip_group_name"` // IP 组名称
}

// 获取CDN记录 请求结构
type GetCdnRecordRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // CDN 记录序号
}

// 创建CDN记录 请求结构
type CreateCdnRecordRequest struct {
	ConfigID       int64  `json:"config_id" binding:"required"`       // 平台配置序号
	Domain         string `json:"domain" binding:"required"`          // 域名（如 www.example.com）
	Type           string `json:"type" binding:"required"`            // 记录类型（A、AAAA等）
	OriginProtocol string `json:"origin_protocol" binding:"required"` // 回源协议（http、https、follow）
	HTTPPort       int    `json:"http_port" binding:"required"`       // HTTP 回源端口
	HTTPSPort      int    `json:"https_port" binding:"required"`      // HTTPS 回源端口
	IPGroupID      int64  `json:"ip_group_id" binding:"required"`     // IP 组序号
}

// 更新CDN记录 请求结构
type UpdateCdnRecordRequest struct {
	GetCdnRecordRequest
	OriginProtocol string `json:"origin_protocol" binding:"required"` // 回源协议（http、https、follow）
	HTTPPort       int    `json:"http_port" binding:"required"`       // HTTP 回源端口
	HTTPSPort      int    `json:"https_port" binding:"required"`      // HTTPS 回源端口
	IPGroupID      int64  `json:"ip_group_id" binding:"required"`     // IP 组序号
}

// 更新CDN记录状态 请求结构
type UpdateCdnRecordActiveRequest struct {
	GetCdnRecordRequest
	IsActive bool `json:"is_active,omitempty"` // 是否活跃
}

// 删除CDN记录 请求结构
type DeleteCdnRecordRequest struct {
	GetCdnRecordRequest
}
