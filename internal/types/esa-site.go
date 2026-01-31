package types

import (
	"vibe_ip/internal/ent"
)

// ESA页面 - 站点记录 请求结构
type ESASiteRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // ESA 站点序号
}

// ESA页面 - 站点记录  响应结构
type ESASiteRecordResponse struct {
	CommonPageWithItemResponse[ent.EsaSite, any, ESASiteRecordRequest]
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
