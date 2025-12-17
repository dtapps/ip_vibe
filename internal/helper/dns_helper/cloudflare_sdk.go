package dns_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

type CloudflareClientSdk struct {
	config struct {
		ApiKey string // API 密钥
		ZoneID string // 站点 ID
	}

	client *cloudflare.Client
}

// Cloudflare
// https://developers.cloudflare.com/api/go/
func NewCloudflareClientSdk(ctx context.Context, apiKey string, zoneID string) (c *CloudflareClientSdk, err error) {

	c = &CloudflareClientSdk{}
	c.config.ApiKey = apiKey
	c.config.ZoneID = zoneID

	c.client = cloudflare.NewClient(
		option.WithAPIToken(apiKey),
	)

	return c, nil
}

// 设置调试模式
func (c *CloudflareClientSdk) SetDebug(debug bool) *CloudflareClientSdk {
	// c.client.WithDebug()
	return c
}

// 查询域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
// domainName: 主域名
// subDomain: 子域名
// recordValue: 记录值
func (c *CloudflareClientSdk) QueryDomain(ctx context.Context, domainName string, subDomain string, recordValue string) (QueryDomainResponse, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, recordValue)
	if err != nil {
		return QueryDomainResponse{}, err
	}

	// 请求
	requestParams := dns.RecordListParams{}
	requestParams.ZoneID = cloudflare.String(c.config.ZoneID)
	requestParams.Name = cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.String(fullDomain)})
	if recordType != "" {
		requestParams.Type = cloudflare.F(dns.RecordListParamsType(recordType))
	}
	page, err := c.client.DNS.Records.List(ctx, requestParams)
	if err != nil {
		return QueryDomainResponse{}, err
	}

	for _, item := range page.Result {
		if item.Name == fullDomain {
			if recordType != "" && item.Type != dns.RecordResponseType(recordType) {
				continue
			}
			return QueryDomainResponse{
				DomainID:    item.ID,
				DomainName:  item.Name,
				RecordType:  string(item.Type),
				RecordValue: item.Content,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// domainName: 主域名
// subDomain: 子域名
// recordValue: 记录值
func (c *CloudflareClientSdk) CreateDomain(ctx context.Context, domainName string, subDomain string, recordValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, recordValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(recordValue)
	_, err = c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
		Body:   requestParams,
	})
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/edit/
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// recordValue: 记录值
func (c *CloudflareClientSdk) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, recordValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, recordValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(recordValue)
	_, err = c.client.DNS.Records.Edit(ctx, domainID, dns.RecordEditParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
		Body:   requestParams,
	})
	if err != nil {
		return err
	}

	return nil
}
