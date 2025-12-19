package dns_helper

import (
	"context"
	"fmt"
	"net/http"

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

	debug  bool // 是否开启调试模式
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
func (c *CloudflareClient) SetDebug(debug bool) *CloudflareClient {
	c.debug = debug
	return c
}

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
// fullDomain: 完整域名
// subDomain: 子域名
// recordType: 记录类型
func (c *CloudflareClient) QueryDomain(ctx context.Context, fullDomain string, subDomain string, recordType string) (QueryDomainResponse, error) {

	// 请求
	requestParams := dns.RecordListParams{}
	requestParams.ZoneID = cloudflare.String(c.config.ZoneID)
	requestParams.Name = cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.String(fullDomain)})
	if recordType != "" {
		requestParams.Type = cloudflare.F(dns.RecordListParamsType(recordType))
	}
	response, err := c.client.DNS.Records.List(ctx, requestParams)
	if err != nil {
		return QueryDomainResponse{}, err
	}
	if c.debug {
		fmt.Printf("查询域名记录 请求参数: %+v\n", requestParams)
		fmt.Printf("查询域名记录 响应参数: %+v\n", response)
	}

	for _, item := range response.Result {
		if item.Name == fullDomain {
			if recordType != "" && item.Type != dns.RecordResponseType(recordType) {
				continue
			}
			return QueryDomainResponse{
				RecordID:    item.ID,
				FullDomain:  item.Name,
				RecordType:  string(item.Type),
				OriginValue: item.Content,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *CloudflareClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(originValue)
	response, err := c.client.DNS.Records.New(ctx, dns.RecordNewParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
		Body:   requestParams,
	})
	if err != nil {
		return err
	}
	if c.debug {
		fmt.Printf("创建域名记录 请求参数: %+v\n", requestParams)
		fmt.Printf("创建域名记录 响应参数: %+v\n", response)
	}

	return nil
}

// 更新域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/edit/
// domainID: 域名ID
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *CloudflareClient) UpdateDomain(ctx context.Context, domainID string, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := dns.ARecordParam{}
	requestParams.Name = cloudflare.String(fullDomain)
	requestParams.Type = cloudflare.F(dns.ARecordType(recordType))
	requestParams.Content = cloudflare.String(originValue)
	response, err := c.client.DNS.Records.Edit(ctx, domainID, dns.RecordEditParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
		Body:   requestParams,
	})
	if err != nil {
		return err
	}
	if c.debug {
		fmt.Printf("更新域名记录 请求参数: %+v\n", requestParams)
		fmt.Printf("更新域名记录 响应参数: %+v\n", response)
	}

	return nil
}
