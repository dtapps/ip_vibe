package dns_helper

import (
	"context"
	"fmt"
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

// 云解析 DNS
// DnsPod
// https://cloud.tencent.com/document/product/302
// https://cloud.tencent.com/document/product/1427
func NewDnsPodClientSdk(ctx context.Context, secretID string, secretKey string) (*DnsPodClientSdk, error) {

	var err error

	c := &DnsPodClientSdk{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey

	// 密钥信息
	c.credential = common.NewCredential(
		secretID,
		secretKey,
	)

	// 实例化client选项
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	// 实例化client对象
	c.client, err = dnspod.NewClient(c.credential, "", cpf)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *DnsPodClientSdk) SetDebug(debug bool) *DnsPodClientSdk {
	c.client.WithDebug(debug)
	return c
}

// 获取域名的解析记录列表
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordList
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) QueryDomain(ctx context.Context, domainName string, domainOrigin string) (DnsPodQueryResponse, error) {

	// 解析域名
	_, domainSub, domainMain, err := utils.ExtractDomainParts(domainName)
	if err != nil {
		return DnsPodQueryResponse{}, err
	}

	// 解析IP
	var domainType = ""
	if domainOrigin != "" {
		domainType, err = utils.GetRecordType(domainOrigin)
		if err != nil {
			return DnsPodQueryResponse{}, err
		}
	}

	// 请求
	requestParams := dnspod.NewDescribeRecordListRequest()
	requestParams.Domain = common.StringPtr(domainMain)   // 域名
	requestParams.Subdomain = common.StringPtr(domainSub) // 解析记录的主机头
	if domainType != "" {
		requestParams.RecordType = common.StringPtr(domainType) // 获取某种类型的解析记录
	}
	response, err := c.client.DescribeRecordList(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return DnsPodQueryResponse{}, fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return DnsPodQueryResponse{}, err
	}

	for _, item := range response.Response.RecordList {
		if *item.Name == domainSub {
			if domainType != "" && *item.Type != domainType {
				continue
			}
			return DnsPodQueryResponse{
				DomainID:     int64(*item.RecordId),
				DomainName:   *item.Name,
				DomainType:   *item.Type,
				DomainOrigin: *item.Value,
			}, nil
		}
	}

	return DnsPodQueryResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=CreateRecord
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) CreateDomain(ctx context.Context, domainName string, domainOrigin string) error {

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
	requestParams := dnspod.NewCreateRecordRequest()
	requestParams.Domain = common.StringPtr(domainMain)     // 域名
	requestParams.RecordType = common.StringPtr(domainType) // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")       // 记录线路
	requestParams.Value = common.StringPtr(domainOrigin)    // 记录值
	requestParams.SubDomain = common.StringPtr(domainSub)   // 主机记录
	_, err = c.client.CreateRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=ModifyRecord
// domainID: 域名ID
// domainName: 域名名称
// domainOrigin: 记录类型值
func (c *DnsPodClientSdk) UpdateDomain(ctx context.Context, domainID int64, domainName string, domainOrigin string) error {

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
	requestParams := dnspod.NewModifyRecordRequest()
	requestParams.Domain = common.StringPtr(domainMain)         // 域名
	requestParams.RecordType = common.StringPtr(domainType)     // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")           // 记录线路
	requestParams.Value = common.StringPtr(domainOrigin)        // 记录值
	requestParams.RecordId = common.Uint64Ptr(uint64(domainID)) // 记录 ID
	requestParams.SubDomain = common.StringPtr(domainSub)       // 主机记录
	_, err = c.client.ModifyRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %s", err)
	}
	if err != nil {
		return err
	}

	return nil
}
