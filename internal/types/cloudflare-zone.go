package types

import (
	"vibe_ip/internal/ent"
)

// Cloudflare页面 - 站点 请求结构
type CloudflareZoneRequest struct {
	CommonPageRequest
}

// Cloudflare页面 - 站点 响应结构
type CloudflareZoneResponse struct {
	CommonPageResponse[GetCloudflareZoneResponse, CloudflareZoneRequest]
}

// Cloudflare页面 - 站点记录 请求结构
type CloudflareZoneRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // Cloudflare 站点序号
}

// Cloudflare页面 - 站点记录  响应结构
type CloudflareZoneRecordResponse struct {
	CommonPageWithItemResponse[ent.CloudflareZone, any, CloudflareZoneRecordRequest]
}
