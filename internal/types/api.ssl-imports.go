package types

import "time"

// 获取 SSL 导入 请求结构
type GetSSLImportResponse struct {
	ID          int64     `json:"id,omitempty"`     // SSL导入序号
	Source      string    `json:"source,omitempty"` // 来源
	Domain      string    `json:"domain,omitempty"` // 域名（支持通配符 *.example.com 或精确域名，多个域名之间用逗号分隔）
	Certificate string    `json:"certificate"`      // 证书(PEM格式)
	PrivateKey  string    `json:"private_key"`      // 私钥(PEM格式)
	NotBefore   time.Time `json:"not_before"`       // 证书开始时间
	NotAfter    time.Time `json:"not_after"`        // 证书结束时间
	Status      string    `json:"status,omitempty"` // 证书状态
}

// 导入 SSL 导入 请求结构
type ImportSSLImportRequest struct {
	Certificate string `json:"certificate" binding:"required" jsonschema:"证书(PEM格式)"` // 证书(PEM格式)
	PrivateKey  string `json:"private_key" binding:"required" jsonschema:"私钥(PEM格式)"` // 私钥(PEM格式)
}

// 创建 SSL 导入 请求结构
type CreateSSLImportRequest struct {
	Domain      string `json:"domain" binding:"required" jsonschema:"域名（支持通配符 *.example.com 或精确域名）使用逗号分隔多个域名"` // 域名（支持通配符 *.example.com 或精确域名）使用逗号分隔多个域名
	Certificate string `json:"certificate" binding:"required" jsonschema:"证书(PEM格式)"`                          // 证书(PEM格式)
	PrivateKey  string `json:"private_key" binding:"required" jsonschema:"私钥(PEM格式)"`                          // 私钥(PEM格式)
}

// 删除 SSL 导入 请求结构
type DeleteSSLImportRequest struct {
	GetSSLRecordRequest
}

// 批量删除 SSL 导入 请求结构
type BatchDeleteSSLImportRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"SSL记录的唯一标识符列表"` // SSL记录序号列表
}
