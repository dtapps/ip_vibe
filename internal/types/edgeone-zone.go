package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_stat"
)

// EdgeOne页面 - 站点 请求结构
type EdgeOneZoneRequest struct {
	CommonPageRequest
}

// EdgeOne页面 - 站点 响应结构
type EdgeOneZoneResponse struct {
	CommonPageResponse[GetEdgeOneZoneResponse, EdgeOneZoneRequest]
}

// EdgeOne页面 - 站点分析 请求结构
type EdgeOneZoneAnalysisRequest struct {
	CommonPageRequest
	ID        int64  `uri:"id" json:"-" binding:"required"` // EdgeOne 站点序号
	TimeRange string `form:"range,omitempty" json:"-"`      // 时间范围
}

// EdgeOne页面 - 站点分析  响应结构
type EdgeOneZoneAnalysisResponse struct {
	CommonPageWithItemAndStatResponse[ent_stat.EdgeOneZoneAnalysisData, ent.EdgeOneZone, ent_stat.EdgeOneZoneAnalysisData, EdgeOneZoneAnalysisRequest]
}

// EdgeOne页面 - 站点记录 请求结构
type EdgeOneZoneRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // EdgeOne 站点序号
}

// EdgeOne页面 - 站点记录  响应结构
type EdgeOneZoneRecordResponse struct {
	CommonPageWithItemResponse[ent.EdgeOneZone, any, EdgeOneZoneRecordRequest]
}

// EdgeOne页面 - 站点源组 请求结构
type EdgeOneZoneOriginGroupRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // EdgeOne 站点序号
}

// EdgeOne页面 - 站点源组  响应结构
type EdgeOneZoneOriginGroupResponse struct {
	CommonPageWithItemResponse[ent.EdgeOneZone, ent.EdgeOneZoneOriginGroup, EdgeOneZoneOriginGroupRequest]
}
