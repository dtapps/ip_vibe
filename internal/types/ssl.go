package types

import (
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// SSL页面 请求结构
type SSLRequest struct {
	CommonPageRequest
	ID            int64        `json:"id,omitempty" form:"id,omitempty"`                 // SSL记录序号
	ConfigID      int64        `json:"config_id,omitempty" form:"config_id,omitempty"`   // 平台配置序号
	Domain        string       `json:"domain,omitempty" form:"domain,omitempty"`         // 域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）
	Status        string       `json:"status,omitempty" form:"status,omitempty"`         // 证书状态
	IsActive      NullableBool `json:"is_active" form:"is_active,omitempty"`             // 是否活跃
	IsAutoRenewal NullableBool `json:"is_auto_renewal" form:"is_auto_renewal,omitempty"` // 是否自动续期
}

// SSL页面 响应结构
type SSLResponse struct {
	CommonPageResponse[GetSSLRecordResponse, SSLRequest] // SSL记录列表
}

// SSL详情页面 请求结构
type SSLDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // SSL记录序号
}

// SSL详情页面 响应结构
type SSLDetailResponse struct {
	CertDetail    UtilsParsedCertificateFull      // 证书详情
	Item          GetSSLRecordResponse            // SSL记录
	Historys      []ent_log.SSLCertificateHistory // 证书申请历史记录
	WebhookSubs   []ent.WebhookSubscription       // 通知订阅
	OperationLogs []ent_log.OperationLog          // 操作日志
}

// SSL证书申请历史记录 请求结构
type SSLHistoryRequest struct {
	SSLDetailRequest
	CommonPageRequest
}

// SSL证书申请历史记录 响应结构
type SSLHistoryResponse struct {
	CommonPageResponse[ent_log.SSLCertificateHistory, SSLHistoryRequest]                      // SSL证书申请历史记录列表
	Item                                                                 GetSSLRecordResponse // SSL记录
}
