package types

import (
	"vibe_ip/internal/ent"
)

// EdgeOne页面 - 站点记录 请求结构
type EdgeOneZoneRecordRequest struct {
	CommonPageRequest
	ID int64 `uri:"id" json:"-" binding:"required"` // EdgeOne 站点序号
}

// EdgeOne页面 - 站点记录  响应结构
type EdgeOneZoneRecordResponse struct {
	CommonPageWithItemResponse[ent.EdgeOneZone, any, EdgeOneZoneRecordRequest]
}
