package types

import (
	"time"
)

// 获取全部 SSL 记录 请求结构
type GetSSLRecordsRequest struct {
	IsActive *bool `json:"is_active,omitempty" form:"is_active,omitempty" jsonschema:"指定是否只返回活跃的SSL记录"` // 是否活跃
}

// 获取 SSL 记录 响应结构
type GetSSLRecordResponse struct {
	ID                    int64     `json:"id" jsonschema:"SSL记录序号"`                                       // SSL记录序号
	ConfigID              int64     `json:"config_id" jsonschema:"平台配置序号"`                                 // 平台配置序号
	ConfigName            string    `json:"config_name" jsonschema:"平台配置序号"`                               // 平台配置名称
	Platform              string    `json:"platform" jsonschema:"平台平台名称"`                                  // 平台平台名称
	Domain                string    `json:"domain" jsonschema:"域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）"` // 域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）
	CertDomain            string    `json:"cert_domain" jsonschema:"证书域名"`                                 // 证书域名
	CertURL               string    `json:"cert_url" jsonschema:"证书URL"`                                   // 证书URL
	CertStableURL         string    `json:"cert_stable_url" jsonschema:"稳定证书URL"`                          // 稳定证书URL
	CertPrivateKey        string    `json:"cert_private_key" jsonschema:"私钥(PEM格式)"`                       // 私钥(PEM格式)
	CertCertificate       string    `json:"cert_certificate" jsonschema:"证书(PEM格式)"`                       // 证书(PEM格式)
	CertIssuerCertificate string    `json:"cert_issuer_certificate" jsonschema:"颁发者证书(PEM格式)"`             // 颁发者证书(PEM格式)
	CertCsr               string    `json:"cert_csr" jsonschema:"证书签名请求(PEM格式)"`                           // 证书签名请求(PEM格式)
	NotBefore             time.Time `json:"not_before" jsonschema:"证书生效时间"`                                // 证书生效时间
	NotAfter              time.Time `json:"not_after" jsonschema:"证书过期时间"`                                 // 证书过期时间
	Status                string    `json:"status" jsonschema:"证书状态"`                                      // 证书状态
	Provider              string    `json:"provider" jsonschema:"CA提供商"`                                   // CA提供商
	IsActive              bool      `json:"is_active" jsonschema:"是否活跃"`                                   // 是否活跃
	RenewalDays           int       `json:"renewal_days" jsonschema:"过期前多少天开始续期"`                          // 过期前多少天开始续期
	IsAutoRenewal         bool      `json:"is_auto_renewal" jsonschema:"是否自动续期"`                           // 是否自动续期
	LastRefreshAt         time.Time `json:"last_refresh_at" jsonschema:"最后一次刷新成功时间"`                       // 最后一次刷新成功时间

	Webhooks []int64 `json:"webhooks" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 获取 SSL 记录 请求结构
type GetSSLRecordRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"SSL记录序号"` // SSL记录序号
}

// 创建 SSL 记录 请求结构
type CreateSSLRecordRequest struct {
	ConfigID    int64  `json:"config_id" binding:"required" jsonschema:"平台配置序号"`                               // 平台配置序号
	Domain      string `json:"domain" binding:"required" jsonschema:"域名（支持通配符 *.example.com 或精确域名）使用逗号分隔多个域名"` // 域名（支持通配符 *.example.com 或精确域名）使用逗号分隔多个域名
	Provider    string `json:"provider" binding:"required" jsonschema:"CA提供商"`                                 // CA提供商
	RenewalDays int    `json:"renewal_days" jsonschema:"过期前多少天开始续期"`                                           // 过期前多少天开始续期

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新 SSL 记录 请求结构
type UpdateSSLRecordRequest struct {
	GetSSLRecordRequest
	Provider    string `json:"provider" binding:"required" jsonschema:"CA提供商"` // CA提供商
	RenewalDays int    `json:"renewal_days" jsonschema:"过期前多少天开始续期"`           // 过期前多少天开始续期

	WebhookIDs []int64 `json:"webhook_ids,omitempty" jsonschema:"关联的Webhook ID列表"` // 关联的 Webhook 列表
}

// 更新 SSL 记录自动续期状态 请求结构
type UpdateSSLRecordAutoRenewalRequest struct {
	GetSSLRecordRequest
	IsAutoRenewal bool `json:"is_auto_renewal,omitempty" jsonschema:"是否自动续期"` // 是否自动续期
}

// 更新 SSL 记录启用状态 请求结构
type UpdateSSLRecordActiveRequest struct {
	GetSSLRecordRequest
	IsActive bool `json:"is_active,omitempty" jsonschema:"是否活跃"` // 是否活跃
}

// 删除 SSL 记录 请求结构
type DeleteSSLRecordRequest struct {
	GetSSLRecordRequest
}

// 申请新的SSL证书 请求结构
type ApplySSLRecordRequest struct {
	GetSSLRecordRequest
}

// 获取SSL记录历史详情 请求结构
type GetSSLRecordHistoryDetailRequest struct {
	GetSSLRecordRequest
	HistoryID int64 `uri:"history_id" json:"history_id" binding:"required" jsonschema:"SSL记录历史序号"` // SSL记录历史序号
}

// 获取SSL记录历史详情 响应结构
type GetSSLRecordHistoryDetailResponse struct {
	ID                    int64     `json:"id"`                      // SSL证书申请历史记录序号
	CreatedAt             time.Time `json:"created_at"`              // 创建时间
	SslRecordID           int64     `json:"ssl_record_id"`           // 关联 SSLRecord 序号
	CertDomain            string    `json:"cert_domain"`             // 证书域名
	CertURL               string    `json:"cert_url"`                // 证书URL
	CertStableURL         string    `json:"cert_stable_url"`         // 稳定证书URL
	CertPrivateKey        string    `json:"cert_private_key"`        // 私钥(PEM格式)
	CertCertificate       string    `json:"cert_certificate"`        // 证书(PEM格式)
	CertIssuerCertificate string    `json:"cert_issuer_certificate"` // 颁发者证书(PEM格式)
	CertCsr               string    `json:"cert_csr"`                // 证书签名请求(PEM格式)
	NotBefore             time.Time `json:"not_before"`              // 证书生效时间
	NotAfter              time.Time `json:"not_after"`               // 证书过期时间
	Provider              string    `json:"provider"`                // CA提供商
}
