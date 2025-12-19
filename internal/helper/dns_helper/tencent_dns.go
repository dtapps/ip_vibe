package dns_helper

import (
	"context"
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"go.dtapp.net/library/contrib/http_log"
)

type TencentDnsClient struct {
	config struct {
		SecretID  string // 密钥ID
		SecretKey string // 密钥密钥
	}

	debug        bool // 是否开启调试模式
	credential   *common.Credential
	clientConfig *profile.ClientProfile
	client       *dnspod.Client
}

// 云解析 DNS
// DnsPod
// https://cloud.tencent.com/document/product/302
// https://cloud.tencent.com/document/product/1427
func NewTencentDnsClient(ctx context.Context, secretID string, secretKey string) (*TencentDnsClient, error) {

	var err error

	c := &TencentDnsClient{}
	c.config.SecretID = secretID
	c.config.SecretKey = secretKey

	// 密钥信息
	c.credential = common.NewCredential(
		c.config.SecretID,
		c.config.SecretKey,
	)

	// 实例化client选项
	c.clientConfig = profile.NewClientProfile()
	c.clientConfig.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"

	// 实例化client对象
	c.client, err = dnspod.NewClient(c.credential, "", c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *TencentDnsClient) SetDebug(debug bool) *TencentDnsClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *TencentDnsClient) SetLog(log *http_log.LoggingTransport) *TencentDnsClient {
	if log == nil {
		return c
	}
	c.client.WithHttpTransport(log)
	return c
}

// 查询域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordList
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
func (c *TencentDnsClient) QueryDomain(ctx context.Context, domainName string, subDomain string, recordType string) (QueryDomainResponse, error) {

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
	if c.debug {
		printJSON("查询域名记录", requestParams, response)
	}

	for _, item := range response.Response.RecordList {
		if tea.StringValue(item.Name) == subDomain {
			if recordType != "" && tea.StringValue(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    cast.ToString(*item.RecordId),
				FullDomain:  tea.StringValue(item.Name),
				RecordType:  tea.StringValue(item.Type),
				OriginValue: tea.StringValue(item.Value),
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=CreateRecord
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
// originValue: 源站地址
func (c *TencentDnsClient) CreateDomain(ctx context.Context, domainName string, subDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := dnspod.NewCreateRecordRequest()
	requestParams.Domain = common.StringPtr(domainName)     // 域名
	requestParams.RecordType = common.StringPtr(recordType) // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")       // 记录线路
	requestParams.Value = common.StringPtr(originValue)     // 记录值
	requestParams.SubDomain = common.StringPtr(subDomain)   // 主机记录
	response, err := c.client.CreateRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建域名记录", requestParams, response)
	}

	return nil
}

// 更新域名记录
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=ModifyRecord
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
// originValue: 源站地址
func (c *TencentDnsClient) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := dnspod.NewModifyRecordRequest()
	requestParams.Domain = common.StringPtr(domainName)                // 域名
	requestParams.RecordType = common.StringPtr(recordType)            // 记录类型
	requestParams.RecordLine = common.StringPtr("默认")                  // 记录线路
	requestParams.Value = common.StringPtr(originValue)                // 记录值
	requestParams.RecordId = common.Uint64Ptr(cast.ToUint64(domainID)) // 记录 ID
	requestParams.SubDomain = common.StringPtr(subDomain)              // 主机记录
	response, err := c.client.ModifyRecord(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新域名记录", requestParams, response)
	}

	return nil
}
