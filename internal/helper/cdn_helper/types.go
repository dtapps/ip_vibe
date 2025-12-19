package cdn_helper

const (
	AliyunEsaTypeAAAAA = "A/AAAA"
	AliyunEsaTypeCNAME = "CNAME"
)

const (
	AliyunEsaSourceTypeDomain = "Domain"
	AliyunEsaSourceTypeOP     = "OP"
)

const (
	OriginProtocolHTTP   = "HTTP"
	OriginProtocolHTTPS  = "HTTPS"
	OriginProtocolFOLLOW = "FOLLOW"
)

// 查询加速域名
type TencentEOQueryDomainResponse struct {
	RecordID        string // 域名ID
	FullDomain      string // 完整域名
	DomainCname     string // 域名别名
	RecordType      string // 记录类型
	OriginValue     string // 源站值
	OriginProtocol  string // 回源协议
	HttpOriginPort  int    // HTTP回源端口
	HttpsOriginPort int    // HTTPS回源端口
}

// 查询源站组
type TencentEOQueryOriginGroupResponse struct {
	GroupID      string   // 源站组ID
	GroupName    string   // 源站组名称
	GroupType    string   // 源站组类型
	GroupOrigins []string // 源站组值列表
}

// 查询加速域名
type AliyunEsaQueryDomainResponse struct {
	RecordID    string // 域名ID
	FullDomain  string // 完整域名
	DomainCname string // 域名别名
	RecordType  string // 记录类型
	OriginValue string // 源站值
}

// 查询源站组
type AliyunEsaQueryOriginGroupResponse struct {
	GroupID      string   // 源站组ID
	GroupName    string   // 源站组名称
	GroupCname   string   // 源站组别名
	GroupOrigins []string // 源站组值列表
}
