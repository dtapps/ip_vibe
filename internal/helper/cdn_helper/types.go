package cdn_helper

// 查询加速域名列表
type QueryDomainResponse struct {
	DomainID        string // 域名ID
	DomainName      string // 域名名称
	DomainCname     string // 域名别名
	RecordType      string // 记录类型
	RecordValue     string // 记录值
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
