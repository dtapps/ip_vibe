package dns_helper

type QueryDomainResponse struct {
	DomainID    string // 域名ID
	DomainName  string // 域名名称
	OriginType  string // 源站类型
	OriginValue string // 源站地址
}
