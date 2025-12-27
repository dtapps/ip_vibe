package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// 域名监控页面 请求结构
type DomainMonitorRequest struct {
	CommonPageRequest
}

// 域名监控页面 响应结构
type DomainMonitorResponse struct {
	CommonResponseWithRequest                    // 请求对象
	CommonPageResponse[GetDomainMonitorResponse] // 域名监控列表
}

// 域名监控详情页面 请求结构
type DomainMonitorDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // 域名监控记录序号
}

// 域名监控详情页面 响应结构
type DomainMonitorDetailResponse struct {
	CommonResponseWithRequest // 请求对象
	Item                      ent.DomainMonitor
	LatestHistory             ent_log.DomainMonitorHistory
}

// 域名监控页面 请求结构
type DomainMonitorHistoryRequest struct {
	DomainMonitorDetailRequest
	CommonPageRequest
}

// 域名监控页面 响应结构
type DomainMonitorHistoryResponse struct {
	CommonResponseWithRequest // 请求对象
	CommonPageResponse[ent_log.DomainMonitorHistory]
	Item ent.DomainMonitor
}
