package cdn_helper

// 查询加速域名列表
type EdgeOneQueryDomainResponse struct {
	DomainID        string // 域名ID
	DomainName      string // 域名名称
	DomainCname     string // 域名别名
	OriginType      string // 源站类型
	OriginValue     string // 源站值
	OriginProtocol  string // 回源协议
	HttpOriginPort  int    // HTTP回源端口
	HttpsOriginPort int    // HTTPS回源端口
}

// 获取源站组列表
type EdgeOneQueryOriginGroupResponse struct {
	GroupID      string   // 源站组ID
	GroupName    string   // 源站组名称
	GroupType    string   // 源站组类型
	GroupOrigins []string // 源站组类型值
}

// 查询加速域名列表
type EdgeSecurityAccelerationResponseQueryDomain struct {
	DomainID    string // 域名ID
	DomainName  string // 域名名称
	DomainCname string // 域名别名
	OriginType  string // 源站类型
	OriginValue string // 源站值
}
