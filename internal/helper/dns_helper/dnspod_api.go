package dns_helper

import (
	"context"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/tencent_helper"
	"vibe_ip/internal/utils"

	"go.dtapp.net/library/service/tencent"
)

type DnsPodClientApi struct {
	config struct {
		SecretID  string // 密钥ID
		SecretKey string // 密钥密钥
	}
	client *tencent.Client
}

func NewDnsPodClientApi(ctx context.Context, secretID string, secretKey string) (c *DnsPodClientApi, err error) {

	c = &DnsPodClientApi{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey

	c.client, err = tencent.NewClient(
		ctx,
		tencent.WithEndpoint("https://dnspod.tencentcloudapi.com"),
		tencent.WithSecretID(secretID),
		tencent.WithSecretKey(secretKey),
		tencent.WithRestyLog(global.GvaRestyLog),
		tencent.WithDebug(global.GvaEnv.DEBUG),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 开启调试
func (c *DnsPodClientApi) EnableDebug() *DnsPodClientApi {
	c.client.WithDebug()
	return c
}

type ResponseDnsPodQuery struct {
	DomainID     int64  // 域名ID
	DomainName   string // 域名名称
	DomainType   string // 记录类型
	DomainOrigin string // 记录类型值
}

// 获取域名的解析记录列表
// https://cloud.tencent.com/document/api/1427/56166
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientApi) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (resp ResponseDnsPodQuery, err error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return resp, err
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return resp, err
	}

	// 请求
	apiResp := tencent_helper.ResponseDescribeRecordList{}
	_, err = c.client.R().
		SetXTCAction("DescribeRecordList").
		SetXTCVersion("2021-03-23").
		SetBodyMap(map[string]any{
			"Domain":     domainMain,
			"Subdomain":  domainSub,
			"RecordType": domainType,
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return resp, err
	}

	for _, item := range apiResp.Response.RecordList {
		if item.Name == domainSub && item.Type == domainType {
			return ResponseDnsPodQuery{
				DomainID:     item.RecordId,
				DomainName:   item.Name,
				DomainType:   item.Type,
				DomainOrigin: item.Value,
			}, nil
		}
	}

	return resp, nil
}

// 添加记录
// https://cloud.tencent.com/document/api/1427/56180
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientApi) CreateDomain(ctx context.Context, domainName string, domainOrigin string) (err error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}
	if domainSub == "" {
		domainSub = "@"
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	apiResp := tencent_helper.ResponseCreateRecord{}
	_, err = c.client.R().
		SetXTCAction("CreateRecord").
		SetXTCVersion("2021-03-23").
		SetBodyMap(map[string]any{
			"Domain":     domainMain,
			"RecordType": domainType,
			"RecordLine": "默认",
			"Value":      domainOrigin,
			"SubDomain":  domainSub,
		}).
		SetResult(&apiResp).
		SetContext(ctx).
		Post("")
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientApi) UpdateDomain(ctx context.Context, domainID int64, domainName string, domainOrigin string) (err error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return err
	}
	if domainSub == "" {
		domainSub = "@"
	}

	// 解析IP
	domainType, err := utils.GetRecordType(domainOrigin)
	if err != nil {
		return err
	}

	// 请求
	apiResp := tencent_helper.ResponseModifyRecord{}
	_, err = c.client.R().
		SetXTCAction("ModifyRecord").
		SetXTCVersion("2021-03-23").
		SetBodyMap(map[string]any{
			"Domain":     domainMain,
			"RecordType": domainType,
			"RecordLine": "默认",
			"Value":      domainOrigin,
			"RecordId":   domainID,
			"SubDomain":  domainSub,
		}).
		SetResult(&apiResp).
		Post("")
	if err != nil {
		return err
	}

	return nil
}
