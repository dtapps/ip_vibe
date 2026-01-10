package schematype

const (
	CDNRecordOriginProtocolHTTP   string = "http"   // HTTP协议
	CDNRecordOriginProtocolHTTPS  string = "https"  // HTTPS协议
	CDNRecordOriginProtocolFollow string = "follow" // 跟随协议
)

// CDNRecordConfigJSON 是 CDN 记录的配置 JSON 结构
type CDNRecordConfigJSON struct {
	OriginProtocol string `json:"origin_protocol"` // 回源协议（http、https、follow）
	HTTPPort       int    `json:"http_port"`       // HTTP 回源端口
	HTTPSPort      int    `json:"https_port"`      // HTTPS 回源端口
}
