package dns_helper

import (
	"context"
	"vibe_ip/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type DnsPodClientSdk struct {
	config struct {
		SecretID  string // 密钥ID
		SecretKey string // 密钥密钥
	}

	credential *common.Credential
	client     *dnspod.Client
}

func NewDnsPodClientSdk(ctx context.Context, secretID string, secretKey string) (c *DnsPodClientSdk, err error) {

	c = &DnsPodClientSdk{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey

	// 密钥信息
	c.credential = common.NewCredential(
		secretID,
		secretKey,
	)

	// 实例化一个client选项
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	c.client, err = dnspod.NewClient(c.credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 开启调试
func (c *DnsPodClientSdk) EnableDebug() *DnsPodClientSdk {
	c.client.WithDebug(true)
	return c
}

// 获取域名的解析记录列表
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordList
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (resp ResponseDnsPodQuery, err error) {

	// 准备请求参数
	requestParams := dnspod.NewDescribeRecordListRequest()

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return resp, err
	}
	requestParams.Domain = common.StringPtr(domainMain)
	requestParams.Subdomain = common.StringPtr(domainSub)

	// 解析IP
	var domainType = ""
	if domainOrigin != "" {
		domainType, err = utils.GetRecordType(domainOrigin)
		if err != nil {
			return resp, err
		}
		requestParams.RecordType = common.StringPtr(domainType)
	}

	// 请求
	response, err := c.client.DescribeRecordList(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		return
	}

	for _, item := range response.Response.RecordList {
		if *item.Name == domainSub {
			if domainType != "" && *item.Type != domainType {
				continue
			}
			return ResponseDnsPodQuery{
				DomainID:     int64(*item.RecordId),
				DomainName:   *item.Name,
				DomainType:   *item.Type,
				DomainOrigin: *item.Value,
			}, nil
		}
	}

	return resp, nil
}

// 添加记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=CreateRecord
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) CreateDomain(ctx context.Context, domainName string, domainOrigin string) (err error) {

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
	request := dnspod.NewCreateRecordRequest()
	request.Domain = common.StringPtr(domainMain)
	request.RecordType = common.StringPtr(domainType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(domainOrigin)
	request.SubDomain = common.StringPtr(domainSub)
	_, err = c.client.CreateRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		return
	}

	return nil
}

// 更新域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=ModifyRecord
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) UpdateDomain(ctx context.Context, domainID int64, domainName string, domainOrigin string) (err error) {

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
	request := dnspod.NewModifyRecordRequest()
	request.Domain = common.StringPtr(domainMain)
	request.RecordType = common.StringPtr(domainType)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(domainOrigin)
	request.RecordId = common.Uint64Ptr(uint64(domainID))
	request.SubDomain = common.StringPtr(domainSub)
	_, err = c.client.ModifyRecord(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return
	}
	if err != nil {
		return
	}

	return nil
}
