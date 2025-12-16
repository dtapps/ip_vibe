package dns_helper

import (
	"context"
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
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientSdk) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (resp ResponseCloudflareQueryDomain, err error) {

	// 准备请求参数
	requestParams := dns.RecordListParams{
		ZoneID: cloudflare.String(c.config.ZoneID),
	}

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return resp, err
	}
	requestParams.Name = cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.String(domainFull)})

	// 解析IP
	var domainType = ""
	if domainOrigin != "" {
		domainType, err = utils.GetRecordType(domainOrigin)
		if err != nil {
			return resp, err
		}
		requestParams.Type = cloudflare.F(dns.RecordListParamsType(domainType))
	}

	// 请求
	page, err := c.client.DNS.Records.List(ctx, requestParams)
	if err != nil {
		return resp, err
	}

	for _, item := range page.Result {
		if item.Name == domainFull {
			if domainType != "" && item.Type != dns.RecordResponseType(domainType) {
				continue
			}
			return ResponseCloudflareQueryDomain{
				DomainID:     item.ID,
				DomainName:   item.Name,
				DomainType:   string(item.Type),
				DomainOrigin: item.Content,
			}, nil
		}
	}

	return resp, nil
}

// 创建域名记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientSdk) CreateDomain(ctx context.Context, domainName string, domainOrigin string) (err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{
		Name:    cloudflare.String(domainFull),
		Type:    cloudflare.F(dns.ARecordType(domainType)),
		Content: cloudflare.String(domainOrigin),
	}
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
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientSdk) UpdateDomain(ctx context.Context, domainID string, domainName string, domainOrigin string) (err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dns.ARecordParam{
		Name:    cloudflare.String(domainFull),
		Type:    cloudflare.F(dns.ARecordType(domainType)),
		Content: cloudflare.String(domainOrigin),
	}
	_, err = c.client.DNS.Records.Edit(
		ctx,
		domainID,
		dns.RecordEditParams{
			ZoneID: cloudflare.String(c.config.ZoneID),
			Body:   requestParams,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
