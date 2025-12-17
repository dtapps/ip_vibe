package dns_helper

import (
	"context"
	"fmt"
	"net/http"
	"vibe_ip/internal/utils"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"go.dtapp.net/library/contrib/http_log"
)

type CloudflareClient struct {
	config struct {
		ApiKey string // API 密钥
		ZoneID string // 站点 ID
	}

	client *cloudflare.Client
}

// Cloudflare
// https://developers.cloudflare.com/api/go/
func NewCloudflareClient(ctx context.Context, apiKey string, zoneID string) (c *CloudflareClient, err error) {

	c = &CloudflareClient{}
	c.config.ApiKey = apiKey
	c.config.ZoneID = zoneID

	c.client = cloudflare.NewClient(
		option.WithAPIToken(c.config.ApiKey),
	)

	return c, nil
}

// 设置调试模式
// func (c *CloudflareClient) SetDebug(debug bool) *CloudflareClient {
// 	c.client.WithDebug()
// 	return c
// }

// 设置HTTP日志记录器
func (c *CloudflareClient) SetLog(log *http_log.LoggingTransport) *CloudflareClient {
	if log == nil {
		return c
	}
	c.client = cloudflare.NewClient(
		option.WithAPIToken(c.config.ApiKey),
		option.WithHTTPClient(&http.Client{
			Transport: log,
		}),
	)
	return c
}

// 查询域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *CloudflareClient) QueryDomain(ctx context.Context, domainName string, subDomain string, originValue string) (QueryDomainResponse, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
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
				OriginType:  string(item.Type),
				OriginValue: item.Content,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *CloudflareClient) CreateDomain(ctx context.Context, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(originValue)
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
// originValue: 源站地址
func (c *CloudflareClient) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(originValue)
	_, err = c.client.DNS.Records.Edit(ctx, domainID, dns.RecordEditParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
		Body:   requestParams,
	})
	if err != nil {
		return err
	}

	return nil
}
