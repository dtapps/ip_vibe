package dns_helper

import (
	"context"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/cloudflare_helper"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/service/cloudflare"
)

type CloudflareClientApi struct {
	config struct {
		ApiKey string // API 密钥
		ZoneID string // 站点 ID
	}
	client *cloudflare.Client
}

func NewCloudflareClientApi(ctx context.Context, apiKey string, zoneID string) (c *CloudflareClientApi, err error) {

	c = &CloudflareClientApi{}
	c.config.ApiKey = apiKey
	c.config.ZoneID = zoneID

	c.client, err = cloudflare.NewClient(
		ctx,
		cloudflare.WithApiKey(c.config.ApiKey),
		cloudflare.WithRestyLog(global.GvaRestyLog),
		cloudflare.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 开启调试
func (c *CloudflareClientApi) EnableDebug() *CloudflareClientApi {
	c.client.WithDebug()
	return c
}

type ResponseCloudflareQueryDomain struct {
	DomainID     string // 域名ID
	DomainName   string // 域名名称
	DomainType   string // 记录类型
	DomainOrigin string // 记录类型值
}

// 查询域名记录
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientApi) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (resp ResponseCloudflareQueryDomain, err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return resp, err
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return resp, err
	}

	// 请求
	apiResp := cloudflare_helper.ResponseDnsRecordsGet{}
	_, err = c.client.R().
		SetQueryParams(map[string]string{
			"name": domainFull,
			"type": domainType,
		}).
		SetPathParams(map[string]string{
			"zone_id": c.config.ZoneID,
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Get("/zones/{zone_id}/dns_records")
	if err != nil {
		return resp, err
	}

	for _, item := range apiResp.Result {
		if item.Name == domainFull && item.Type == domainType {
			return ResponseCloudflareQueryDomain{
				DomainID:     item.ID,
				DomainName:   item.Name,
				DomainType:   item.Type,
				DomainOrigin: item.Content,
			}, nil
		}
	}

	return resp, nil
}

// 创建域名记录
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientApi) CreateDomain(ctx context.Context, domainName string, domainOrigin string) (err error) {

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
	_, err = c.client.R().
		SetBodyMap(map[string]any{
			"name":    domainFull,
			"type":    domainType,
			"content": domainOrigin,
		}).
		SetPathParams(map[string]string{
			"zone_id": c.config.ZoneID,
		}).
		SetContext(ctx).
		Post("/zones/{zone_id}/dns_records")
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *CloudflareClientApi) UpdateDomain(ctx context.Context, domainID string, domainName string, domainOrigin string) (err error) {

	// 解析域名
	domainFull, _, _, err := utils.ExtractDomainParts(domainName)

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	_, err = c.client.R().
		SetBodyMap(map[string]any{
			"name":    domainFull,
			"type":    domainType,
			"content": domainOrigin,
		}).
		SetPathParams(map[string]string{
			"zone_id":       c.config.ZoneID,
			"dns_record_id": domainID,
		}).
		SetContext(ctx).
		Patch("/zones/{zone_id}/dns_records/{dns_record_id}")
	if err != nil {
		return err
	}

	return nil
}
