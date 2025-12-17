package dns_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/cast"
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
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *DnsPodClientSdk) QueryDomain(ctx context.Context, domainName string, subDomain string, originValue string) (QueryDomainResponse, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return QueryDomainResponse{}, err
	}

	// 请求
	requestParams := dnspod.NewDescribeRecordListRequest()
	requestParams.Domain = common.StringPtr(domainName)   // 域名
	requestParams.Subdomain = common.StringPtr(subDomain) // 解析记录的主机头
	if recordType != "" {
		requestParams.RecordType = common.StringPtr(recordType) // 获取某种类型的解析记录
	}
	response, err := c.client.DescribeRecordList(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return QueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return QueryDomainResponse{}, err
	}

	for _, item := range response.Response.RecordList {
		if tea.StringValue(item.Name) == subDomain {
			if recordType != "" && tea.StringValue(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				DomainID:    tea.ToString(item.RecordId),
				DomainName:  tea.StringValue(item.Name),
				OriginType:  tea.StringValue(item.Type),
				OriginValue: tea.StringValue(item.Value),
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=CreateRecord
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *DnsPodClientSdk) CreateDomain(ctx context.Context, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dnspod.NewCreateRecordRequest()
	requestParams.Domain = common.StringPtr(domainName)     // 域名
	requestParams.RecordType = common.StringPtr(recordType) // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")       // 记录线路
	requestParams.Value = common.StringPtr(originValue)     // 记录值
	requestParams.SubDomain = common.StringPtr(subDomain)   // 主机记录
	_, err = c.client.CreateRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=ModifyRecord
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *DnsPodClientSdk) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := dnspod.NewModifyRecordRequest()
	requestParams.Domain = common.StringPtr(domainName)                // 域名
	requestParams.RecordType = common.StringPtr(recordType)            // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")                  // 记录线路
	requestParams.Value = common.StringPtr(originValue)                // 记录值
	requestParams.RecordId = common.Uint64Ptr(cast.ToUint64(domainID)) // 记录 ID
	requestParams.SubDomain = common.StringPtr(subDomain)              // 主机记录
	_, err = c.client.ModifyRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}
