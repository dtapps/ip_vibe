package types

import (
	"vibe_ip/internal/ent_log"
)

// SSL导入页面 请求结构
type SSLImportRequest struct {
	CommonPageRequest
}

// SSL导入页面 响应结构
type SSLImportResponse struct {
	CommonPageResponse[GetSSLImportResponse, SSLImportRequest] // SSL记录列表
}

// SSL导入详情页面 请求结构
type SSLImportDetailRequest struct {
	ID int64 `uri:"id" json:"-" binding:"required"` // SSL导入记录序号
}

// SSL导入详情页面 响应结构
type SSLImportDetailResponse struct {
	CertDetail    UtilsParsedCertificateFull // 证书详情
	Item          GetSSLImportResponse       // SSL导入记录
	OperationLogs []ent_log.OperationLog     // 操作日志
}
