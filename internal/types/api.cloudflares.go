package types

// 获取Cloudflare域名 请求结构
type GetCloudflareDomainRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"平台配置序号"` // 平台配置序号
}

// 获取Cloudflare域名 入站配置
type GetCloudflareDomainResponse struct {
	ZoneID   string `json:"zone_id"`   // 区域ID
	ZoneName string `json:"zone_name"` // 区域名称
}
