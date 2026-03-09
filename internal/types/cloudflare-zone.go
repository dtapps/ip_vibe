package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_stat"
)

// Cloudflare页面 - 站点 请求结构
type CloudflareZoneRequest struct {
	CommonPageRequest
}

// Cloudflare页面 - 站点 响应结构
type CloudflareZoneResponse struct {
	CommonPageResponse[GetCloudflareZoneResponse, CloudflareZoneRequest]
}

// Cloudflare页面 - 站点分析 请求结构
type CloudflareZoneAnalysisRequest struct {
	CommonPageRequest
	ID        int64  `uri:"id" json:"-" binding:"required"` // Cloudflare 站点序号
	TimeRange string `form:"range,omitempty" json:"-"`      // 时间范围
}

// Cloudflare页面 - 站点分析  响应结构
type CloudflareZoneAnalysisResponse struct {
	CommonPageWithItemAndStatResponse[ent_stat.CloudflareZoneAnalysisData, ent.CloudflareZone, ent_stat.CloudflareZoneAnalysisData, CloudflareZoneAnalysisRequest]
}

// Cloudflare页面 - 站点记录 请求结构
type CloudflareZoneRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // Cloudflare 站点序号
}

// Cloudflare页面 - 站点记录  响应结构
type CloudflareZoneRecordResponse struct {
	CommonPageWithItemResponse[ent.CloudflareZone, ent.CloudflareZoneRecord, CloudflareZoneRecordRequest]
}
