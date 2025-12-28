package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// SSL页面 请求结构
type SSLRequest struct {
	CommonPageRequest
}

// SSL页面 响应结构
type SSLResponse struct {
	CommonResponseWithRequest                // 请求对象
	CommonPageResponse[GetSSLRecordResponse] // SSL记录列表
}

// SSL详情页面 请求结构
type SSLDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // SSL记录序号
}

// SSL详情页面 响应结构
type SSLDetailResponse struct {
	CommonResponseWithRequest       // 请求对象
	ExpiryDays                int64 // 证书过期天数
	Item                      ent.SSLRecord
	Historys                  []ent_log.SSLCertificateHistory
}

// SSL证书申请历史记录 请求结构
type SSLHistoryRequest struct {
	SSLDetailRequest
	CommonPageRequest
}

// SSL证书申请历史记录 响应结构
type SSLHistoryResponse struct {
	CommonResponseWithRequest // 请求对象
	CommonPageResponse[ent_log.SSLCertificateHistory]
	Item ent.SSLRecord
}
