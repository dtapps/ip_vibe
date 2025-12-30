package types

import (
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
)

// SSL页面 请求结构
type SSLRequest struct {
	CommonPageRequest
}

// SSL页面 响应结构
type SSLResponse struct {
	CommonPageResponse[GetSSLRecordResponse, SSLRequest] // SSL记录列表
}

// SSL详情页面 请求结构
type SSLDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // SSL记录序号
}

type SSLDetailCertDetail struct {
	// --- 身份信息 ---
	Subject      string   `json:"subject"`       // 主体 (Common Name)
	Issuer       string   `json:"issuer"`        // 颁发者 (例如: Let's Encrypt / ZeroSSL)
	DNSNames     []string `json:"dns_names"`     // SANs (支持的所有域名，包括通配符)
	SerialNumber string   `json:"serial_number"` // 证书唯一序列号

	// --- 算法与安全 ---
	SignatureAlgorithm string `json:"signature_algorithm"`  // 签名算法 (如 SHA256-RSA)
	PublicKeyAlgorithm string `json:"public_key_algorithm"` // 公钥类型 (如 RSA 2048 或 ECDSA P256)
	KeyLength          int    `json:"key_length"`           // 密钥长度 (如 2048, 4096)
	Version            int    `json:"version"`              // X509 版本

	// --- 时间与寿命 ---
	NotBefore     time.Time `json:"not_before"`     // 证书定义的生效时间
	NotAfter      time.Time `json:"not_after"`      // 证书定义的过期时间
	ExpiryDays    int       `json:"expiry_days"`    // 剩余天数
	ExpiryPercent int       `json:"expiry_percent"` // 生命周期消耗百分比

	// --- 状态标志 (用于 UI 逻辑) ---
	IsExpired  bool `json:"is_expired"`  // 是否已过期
	IsWildcard bool `json:"is_wildcard"` // 是否包含通配符域名
	IsECC      bool `json:"is_ecc"`      // 是否是椭圆曲线加密 (更先进/高性能)
}

// SSL详情页面 响应结构
type SSLDetailResponse struct {
	ExpiryDays    int64                           // 证书过期天数
	CertDetail    SSLDetailCertDetail             // 证书详情
	Item          ent.SSLRecord                   // SSL记录
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
	CommonPageResponse[ent_log.SSLCertificateHistory, SSLHistoryRequest] // SSL证书申请历史记录列表
	Item                                                                 ent.SSLRecord
}
