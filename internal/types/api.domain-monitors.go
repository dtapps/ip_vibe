package types

import (
	"time"
)

// 获取全部域名监控 请求结构
type GetDomainMonitorsRequest struct {
	IsActive *bool `json:"is_active" binding:"required"` // 是否活跃
}

// 获取域名监控 响应结构
type GetDomainMonitorResponse struct {
	ID            int64     `json:"id"`              // 域名监控记录序号
	Name          string    `json:"name"`            // 名称
	Domain        string    `json:"domain"`          // 域名
	CheckInterval int       `json:"check_interval"`  // 检查间隔（秒）
	CheckProtocol string    `json:"check_protocol"`  // 检查类型（HTTP/HTTPS）
	IsActive      bool      `json:"is_active"`       // 是否活跃
	CurrentStatus string    `json:"current_status"`  // 当前运行状态
	LastSuccessAt time.Time `json:"last_success_at"` // 最后一次检查成功时间
	LastFailureAt time.Time `json:"last_failure_at"` // 最后一次检查失败时间

	Webhooks []int64 `json:"webhooks,omitempty"` // 关联的 Webhook 列表
}

// 获取域名监控 请求结构
type GetDomainMonitorRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 域名监控记录序号
}

// 获取域名监控 请求结构
type CreateDomainMonitorRequest struct {
	Name          string `json:"name"`           // 名称
	Domain        string `json:"domain"`         // 域名
	CheckInterval int    `json:"check_interval"` // 检查间隔（秒）
	CheckProtocol string `json:"check_protocol"` // 检查类型（HTTP/HTTPS）

	WebhookIDs []int64 `json:"webhook_ids,omitempty"` // 关联的 Webhook 列表
}

// 更新域名监控 请求结构
type UpdateDomainMonitorRequest struct {
	GetDomainMonitorRequest
	Name          string `json:"name"`           // 名称
	CheckInterval int    `json:"check_interval"` // 检查间隔（秒）

	WebhookIDs []int64 `json:"webhook_ids,omitempty"` // 关联的 Webhook 列表
}

// 更新域名监控启用状态 请求结构
type UpdateDomainMonitorActiveRequest struct {
	GetDomainMonitorRequest
	IsActive bool `json:"is_active,omitempty"` // 是否活跃
}

// 删除域名监控 请求结构
type DeleteDomainMonitorRequest struct {
	GetDomainMonitorRequest
}

// 检查域名监控 请求结构
type CheckDomainMonitorRequest struct {
	GetDomainMonitorRequest
}
