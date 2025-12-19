package dns_helper

type QueryDomainResponse struct {
	RecordID    string // 域名ID
	FullDomain  string // 完整域名
	RecordType  string // 记录类型
	OriginValue string // 源站值
}
