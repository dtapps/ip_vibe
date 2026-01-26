package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// 域名监控页面 请求结构
type DomainMonitorRequest struct {
	CommonPageRequest
	ID            int64        `json:"id,omitempty" form:"id,omitempty"`                         // 域名监控记录序号
	Name          string       `json:"name,omitempty" form:"name,omitempty"`                     // 域名监控名称
	Domain        string       `json:"domain,omitempty" form:"domain,omitempty"`                 // 域名
	CheckProtocol string       `json:"check_protocol,omitempty" form:"check_protocol,omitempty"` // 检查类型（HTTP/HTTPS）
	IsActive      NullableBool `json:"is_active,omitempty" form:"is_active,omitempty"`           // 是否活跃
	CurrentStatus string       `json:"current_status,omitempty" form:"current_status,omitempty"` // 当前运行状态
}

// 域名监控页面 响应结构
type DomainMonitorResponse struct {
	CommonPageResponse[GetDomainMonitorResponse, DomainMonitorRequest] // 域名监控列表
}

// 域名监控详情页面 请求结构
type DomainMonitorDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 域名监控记录序号
}

// 域名监控详情页面 响应结构
type DomainMonitorDetailResponse struct {
	Item          GetDomainMonitorResponse     // 域名监控记录
	LatestHistory ent_log.DomainMonitorHistory // 最新监控记录
	WebhookSubs   []ent.WebhookSubscription    // 通知订阅
	OperationLogs []ent_log.OperationLog       // 最近操作日志
}

// 域名监控页面 请求结构
type DomainMonitorHistoryRequest struct {
	DomainMonitorDetailRequest
	CommonPageRequest
}

// 域名监控页面 响应结构
type DomainMonitorHistoryResponse struct {
	CommonPageResponse[ent_log.DomainMonitorHistory, DomainMonitorHistoryRequest]                          // 域名监控记录列表
	Item                                                                          GetDomainMonitorResponse // 域名监控记录
}
