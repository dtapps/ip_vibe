package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_stat"
)

// ESA页面 - 站点 请求结构
type ESASiteRequest struct {
	CommonPageRequest
}

// ESA页面 - 站点 响应结构
type ESASiteResponse struct {
	CommonPageResponse[GetESASiteResponse, ESASiteRequest]
}

// ESA页面 - 站点分析 请求结构
type ESASiteAnalysisRequest struct {
	CommonPageRequest
	ID        int64  `uri:"id" json:"-" binding:"required"` // ESA 站点序号
	TimeRange string `form:"range,omitempty" json:"-"`      // 时间范围
}

// ESA页面 - 站点分析  响应结构
type ESASiteAnalysisResponse struct {
	CommonPageWithItemAndStatResponse[ent_stat.EsaSiteAnalysisData, ent.EsaSite, ent_stat.EsaSiteAnalysisData, ESASiteAnalysisRequest]
}

// ESA页面 - 站点记录 请求结构
type ESASiteRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // ESA 站点序号
}

// ESA页面 - 站点记录  响应结构
type ESASiteRecordResponse struct {
	CommonPageWithItemResponse[ent.EsaSite, ent.EsaSiteRecord, ESASiteRecordRequest]
}

// ESA页面 - 站点源地址池 请求结构
type ESASiteOriginPoolRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // ESA 站点序号
}

// ESA页面 - 站点源地址池  响应结构
type ESASiteOriginPoolResponse struct {
	CommonPageWithItemResponse[ent.EsaSite, ent.EsaSiteOriginPool, ESASiteOriginPoolRequest]
}
