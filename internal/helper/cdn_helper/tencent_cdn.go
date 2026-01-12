package cdn_helper

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"go.dtapp.net/library/contrib/http_log"
)

type TencentCdnConfig struct {
	SecretID   string `json:"secret_id"`   // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥Key
	DomainName string `json:"domain_name"` // 主域名
}

type TencentCdnClient struct {
	config       *TencentCdnConfig // 配置
	debug        bool              // 是否开启调试模式
	credential   *common.Credential
	clientConfig *profile.ClientProfile
	client       *cdn.Client
}

// 腾讯云 内容分发网络 CDN
// https://cloud.tencent.com/product/cdn
// https://console.cloud.tencent.com/api/explorer
func NewTencentCdnClient(ctx context.Context, config *TencentCdnConfig) (*TencentCdnClient, error) {

	var err error

	c := &TencentCdnClient{}
	c.config = config

	// 密钥信息
	c.credential = common.NewCredential(
		c.config.SecretID,
		c.config.SecretKey,
	)

	// 实例化一个client选项
	c.clientConfig = profile.NewClientProfile()
	c.clientConfig.HttpProfile.Endpoint = "cdn.tencentcloudapi.com"

	// 实例化要请求产品的client对象
	c.client, err = cdn.NewClient(c.credential, "", c.clientConfig)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *TencentCdnClient) SetDebug(debug bool) *TencentCdnClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *TencentCdnClient) SetLog(loggingTransport *http_log.LoggingRoundTripper) *TencentCdnClient {
	if loggingTransport == nil {
		return c
	}
	c.client.WithHttpTransport(loggingTransport)
	return c
}

// 查询加速域名
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=DescribeDomains
// fullDomain: 完整域名
func (c *TencentCdnClient) QueryDomain(ctx context.Context, fullDomain string) (TencentCdnQueryDomainResponse, error) {

	// 请求
	requestParams := cdn.NewDescribeDomainsRequest()
	requestParams.Offset = common.Int64Ptr(0)
	requestParams.Limit = common.Int64Ptr(1000)
	requestParams.Filters = make([]*cdn.DomainFilter, 0)
	requestParams.Filters = append(requestParams.Filters, &cdn.DomainFilter{
		Name:  common.StringPtr("domain"),
		Value: common.StringPtrs([]string{fullDomain}),
		Fuzzy: common.BoolPtr(false),
	})
	response, err := c.client.DescribeDomains(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return TencentCdnQueryDomainResponse{}, fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return TencentCdnQueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询加速域名", requestParams, response)
	}

	for _, item := range response.Response.Domains {
		if strings.EqualFold(utils.PtrToString(item.Domain), fullDomain) {
			return TencentCdnQueryDomainResponse{
				RecordID:       utils.PtrToString(item.ResourceId),
				FullDomain:     utils.PtrToString(item.Domain),
				DomainCname:    utils.TrimDots(utils.PtrToString(item.Cname)),
				RecordType:     utils.PtrToString(item.Origin.OriginType),
				OriginValue:    utils.PtrToString(item.Origin.Origins[0]),
				OriginProtocol: utils.PtrToString(item.Origin.OriginPullProtocol),
			}, nil
		}
	}

	return TencentCdnQueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=AddCdnDomain
// fullDomain: 完整域名
// recordType: domain、domainv6、ipv6、ip_ipv6
// originValue: 源站地址或源站组ID
func (c *TencentCdnClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := cdn.NewAddCdnDomainRequest()
	requestParams.Domain = common.StringPtr(fullDomain) // 域名
	requestParams.ServiceType = common.StringPtr("web") // 加速域名业务类型
	requestParams.Origin = &cdn.Origin{
		Origins:    common.StringPtrs([]string{originValue}), // 源站地址
		OriginType: common.StringPtr(recordType),             // 源站类型
	}
	response, err := c.client.AddCdnDomain(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建加速域名", requestParams, response)
	}

	return nil
}

// 更新加速域名
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=UpdateDomainConfig
// fullDomain: 完整域名
// fullDomain: 完整域名
// recordType: domain、domainv6、ipv6、ip_ipv6
// originValue: 源站地址或源站组ID
func (c *TencentCdnClient) UpdateDomain(ctx context.Context, domainID string, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := cdn.NewUpdateDomainConfigRequest()
	requestParams.Domain = common.StringPtr(fullDomain)               // 域名
	requestParams.ProjectId = common.Int64Ptr(cast.ToInt64(domainID)) // 项目 ID
	requestParams.Origin = &cdn.Origin{
		Origins:    common.StringPtrs([]string{originValue}), // 源站地址
		OriginType: common.StringPtr(recordType),             // 源站类型
	}
	response, err := c.client.UpdateDomainConfig(requestParams)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("API错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新加速域名", requestParams, response)
	}

	return nil
}
