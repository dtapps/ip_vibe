package dns_helper

type QueryDomainResponse struct {
	DomainID    string // 域名ID
	DomainName  string // 域名名称
	RecordType  string // 记录类型
	RecordValue string // 记录值
}
