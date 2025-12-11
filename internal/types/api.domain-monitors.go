package types

import (
	"time"
)

// 获取全部域名监控 请求结构
type GetDomainMonitorsRequest struct {
	IsActive *bool `json:"is_active" binding:"required" jsonschema:"指定是否只返回活跃的域名监控记录"` // 是否活跃
}

// 获取域名监控 响应结构
type GetDomainMonitorResponse struct {
	ID            int64     `json:"id" jsonschema:"域名监控记录的唯一标识符"`                 // 域名监控记录序号
	Name          string    `json:"name" jsonschema:"域名监控的名称"`                    // 名称
	Domain        string    `json:"domain" jsonschema:"要监控的域名"`                   // 域名
	CheckInterval int       `json:"check_interval" jsonschema:"检查间隔（秒）"`          // 检查间隔（秒）
	CheckProtocol string    `json:"check_protocol" jsonschema:"检查类型（HTTP/HTTPS）"` // 检查类型（HTTP/HTTPS）
	IsActive      bool      `json:"is_active" jsonschema:"是否活跃"`                  // 是否活跃
	CurrentStatus string    `json:"current_status" jsonschema:"当前运行状态"`           // 当前运行状态
	LastSuccessAt time.Time `json:"last_success_at" jsonschema:"最后一次检查成功时间"`      // 最后一次检查成功时间
	LastFailureAt time.Time `json:"last_failure_at" jsonschema:"最后一次检查失败时间"`      // 最后一次检查失败时间

	Webhooks []int64 `json:"webhooks,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取域名监控 请求结构
type GetDomainMonitorRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"域名监控记录的唯一标识符"` // 域名监控记录序号
}

// 创建域名监控 请求结构
type CreateDomainMonitorRequest struct {
	Name          string `json:"name" jsonschema:"域名监控的名称"`                    // 名称
	Domain        string `json:"domain" jsonschema:"要监控的域名"`                   // 域名
	CheckInterval int    `json:"check_interval" jsonschema:"检查间隔（秒）"`          // 检查间隔（秒）
	CheckProtocol string `json:"check_protocol" jsonschema:"检查类型（HTTP/HTTPS）"` // 检查类型（HTTP/HTTPS）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新域名监控 请求结构
type UpdateDomainMonitorRequest struct {
	GetDomainMonitorRequest
	Name          string `json:"name" jsonschema:"域名监控的名称"`           // 名称
	CheckInterval int    `json:"check_interval" jsonschema:"检查间隔（秒）"` // 检查间隔（秒）

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新域名监控启用状态 请求结构
type UpdateDomainMonitorActiveRequest struct {
	GetDomainMonitorRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 删除域名监控 请求结构
type DeleteDomainMonitorRequest struct {
	GetDomainMonitorRequest
}

// 检查域名监控 请求结构
type CheckDomainMonitorRequest struct {
	GetDomainMonitorRequest
}
