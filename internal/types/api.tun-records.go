package types

import (
	"encoding/json"
	"time"
)

// 获取全部TUN记录 请求结构
type GetTunRecordsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 获取TUN记录 响应结构
type GetTunRecordResponse struct {
	ID            int64           `json:"id" jsonschema:"TUN记录序号"`                        // TUN记录序号
	ConfigID      int64           `json:"config_id" jsonschema:"关联的平台配置ID"`               // 平台配置序号
	ConfigName    string          `json:"config_name" jsonschema:"平台配置序号"`                // 平台配置名称
	Platform      string          `json:"platform" jsonschema:"平台平台名称"`                   // 平台平台名称
	RecordID      any             `json:"record_id" jsonschema:"平台侧记录ID"`                 // 平台侧记录ID
	SubDomain     string          `json:"sub_domain" jsonschema:"子域名（如 www, blog, api）"`  // 子域名（如 www, blog, api）
	DomainPath    string          `json:"domain_path" jsonschema:"域名路径（如 /api/v1）"`       // 域名路径（如 /api/v1）
	ServeType     string          `json:"serve_type" jsonschema:"服务类型（如 HTTP、HTTPS）"`     // 服务类型（如 HTTP、HTTPS）
	ServeURI      string          `json:"serve_uri" jsonschema:"服务URI（如 127.0.0.1:8080）"` // 服务URI（如 127.0.0.1:8080）
	ConfigJSON    json.RawMessage `json:"config_json" jsonschema:"TUN记录配置（JSON）"`         // TUN记录配置（JSON）
	IsActive      bool            `json:"is_active" jsonschema:"是否活跃"`                    // 是否活跃
	LastRefreshAt time.Time       `json:"last_refresh_at" jsonschema:"最后一次刷新成功时间"`        // 最后一次刷新成功时间

	ConfigJSONAccountID  string `json:"-"` // 平台JSON配置（JSON）中的账号ID
	ConfigJSONTunnelID   string `json:"-"` // TUN记录配置（JSON）中的隧道ID
	ConfigJSONTunnelName string `json:"-"` // TUN记录配置（JSON）中的隧道名称
	ConfigJSONZoneID     string `json:"-"` // TUN记录配置（JSON）中的区域ID
	ConfigJSONZoneName   string `json:"-"` // TUN记录配置（JSON）中的区域名称

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表

	CreatedAt         time.Time `json:"-"` // 创建时间
	UpdatedAt         time.Time `json:"-"` // 更新时间
	LastAPIResponse   string    `json:"-"` // 最近一次接口的反馈
	LastAPIResponseAt time.Time `json:"-"` // 最近一次接口的反馈时间
}

// 获取TUN记录 请求结构
type GetTunRecordRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"TUN记录序号"` // TUN记录序号
}

// 创建TUN记录 请求结构
type CreateTunRecordRequest struct {
	ConfigID   int64  `json:"config_id" binding:"required" jsonschema:"关联的平台配置ID"`               // 平台配置序号
	SubDomain  string `json:"sub_domain" binding:"required" jsonschema:"子域名（如 www, blog, api）"`  // 子域名（如 www, blog, api）
	DomainPath string `json:"domain_path,omitempty" jsonschema:"域名路径（如 /api/v1）"`                // 域名路径（如 /api/v1）
	ServeType  string `json:"serve_type" binding:"required" jsonschema:"服务类型（如 HTTP、HTTPS）"`     // 服务类型（如 HTTP、HTTPS）
	ServeURI   string `json:"serve_uri" binding:"required" jsonschema:"服务URI（如 127.0.0.1:8080）"` // 服务URI（如 127.0.0.1:8080）
	ConfigJSON string `json:"config_json,omitempty" jsonschema:"TUN记录配置（JSON）"`                  // TUN记录配置（JSON）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新TUN记录 请求结构
type UpdateTunRecordRequest struct {
	GetTunRecordRequest
	SubDomain  string `json:"sub_domain" binding:"required" jsonschema:"子域名（如 www, blog, api）"`  // 子域名（如 www, blog, api）
	DomainPath string `json:"domain_path,omitempty" jsonschema:"域名路径（如 /api/v1）"`                // 域名路径（如 /api/v1）
	ServeType  string `json:"serve_type" binding:"required" jsonschema:"服务类型（如 HTTP、HTTPS）"`     // 服务类型（如 HTTP、HTTPS）
	ServeURI   string `json:"serve_uri" binding:"required" jsonschema:"服务URI（如 127.0.0.1:8080）"` // 服务URI（如 127.0.0.1:8080）
	ConfigJSON string `json:"config_json,omitempty" jsonschema:"TUN记录配置（JSON）"`                  // TUN记录配置（JSON）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新TUN记录状态 请求结构
type UpdateTunRecordActiveRequest struct {
	GetTunRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 批量更新TUN记录状态 请求结构
type BatchUpdateTunRecordActiveRequest struct {
	BatchDeleteTunRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 删除TUN记录 请求结构
type DeleteTunRecordRequest struct {
	GetTunRecordRequest
}

// 批量删除TUN记录 请求结构
type BatchDeleteTunRecordRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"TUN记录序号列表"` // TUN记录序号列表
}
