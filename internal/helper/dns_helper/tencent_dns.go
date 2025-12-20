package dns_helper

import (
	"context"
	"fmt"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"go.dtapp.net/library/contrib/http_log"
)

type TencentDnsConfig struct {
	SecretID   string `json:"secret_id"`   // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥Key
	DomainName string `json:"domain_name"` // 主域名
}

type TencentDnsClient struct {
	config       *TencentDnsConfig // 配置
	debug        bool              // 是否开启调试模式
	credential   *common.Credential
	clientConfig *profile.ClientProfile
	client       *dnspod.Client
}

// 云解析 DNS
// DnsPod
// https://cloud.tencent.com/document/product/302
// https://cloud.tencent.com/document/product/1427
func NewTencentDnsClient(ctx context.Context, config *TencentDnsConfig) (*TencentDnsClient, error) {

	var err error

	c := &TencentDnsClient{}
	c.config = config

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
func (c *TencentDnsClient) SetLog(loggingTransport *http_log.LoggingTransport) *TencentDnsClient {
	if loggingTransport == nil {
		return c
	}
	c.client.WithHttpTransport(loggingTransport)
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
		if utils.PtrToString(item.Name) == subDomain {
			if recordType != "" && utils.PtrToString(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    cast.ToString(*item.RecordId),
				FullDomain:  utils.PtrToString(item.Name),
				RecordType:  utils.PtrToString(item.Type),
				OriginValue: utils.PtrToString(item.Value),
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
