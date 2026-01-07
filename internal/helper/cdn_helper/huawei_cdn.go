package cdn_helper

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	cdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
	regionCdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/region"
	"go.dtapp.net/library/contrib/http_log"
)

type HuaweiCdnConfig struct {
	region          string // 区域
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥密钥
	DomainName      string `json:"domain_name"`       // 主域名
}

type HuaweiCdnClient struct {
	config  *HuaweiCdnConfig // 配置
	debug   bool             // 是否开启调试模式
	auth    *auth.BasicCredentials
	region  *region.Region
	builder *core.HcHttpClient
	client  *cdn.CdnClient
}

// 华为云 内容分发网络 CDN
// https://www.huaweicloud.com/product/cdn.html
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/doc
func NewHuaweiCdnClient(ctx context.Context, config *HuaweiCdnConfig) (*HuaweiCdnClient, error) {

	var err error

	c := &HuaweiCdnClient{}
	c.config = config
	c.config.region = "cn-south-1"

	// 认证
	c.auth, err = basic.NewCredentialsBuilder().
		WithAk(c.config.AccessKeyID).
		WithSk(c.config.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	// 区域
	c.region, err = regionCdn.SafeValueOf(c.config.region)
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	c.builder, err = cdn.CdnClientBuilder().
		WithRegion(c.region).
		WithCredential(c.auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}
	c.client = cdn.NewCdnClient(c.builder)

	return c, nil
}

// 设置调试模式
func (c *HuaweiCdnClient) SetDebug(debug bool) *HuaweiCdnClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *HuaweiCdnClient) SetLog(loggingTransport *http_log.LoggingTransport) *HuaweiCdnClient {
	if loggingTransport == nil {
		return c
	}
	// 自定义HTTP传输
	httpConfig := config.DefaultHttpConfig().WithHttpRoundTripper(loggingTransport)
	// 实例化client对象
	builder, err := cdn.CdnClientBuilder().
		WithRegion(c.region).
		WithCredential(c.auth).
		WithHttpConfig(httpConfig).
		SafeBuild()
	if err != nil {
		return c
	}
	c.builder = builder
	c.client = cdn.NewCdnClient(c.builder)
	return c
}

// 查询加速域名
type HuaweiCdnQueryDomainResponse struct {
	RecordID    string // 域名ID
	FullDomain  string // 完整域名
	RecordType  string // 记录类型
	OriginValue string // 源站值
}

// 查询加速域名
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/sdk?version=v2&api=ListDomains
// fullDomain: 完整域名
func (c *HuaweiCdnClient) QueryDomain(ctx context.Context, fullDomain string) (HuaweiCdnQueryDomainResponse, error) {

	// 请求
	requestParams := &model.ListDomainsRequest{}
	requestParams.DomainName = utils.StringToPtr(fullDomain) // 加速域名
	requestParams.PageSize = utils.Int32ToPtr(10000)         // 每页加速域名的数量
	requestParams.PageNumber = utils.Int32ToPtr(1)           // 查询的页码
	requestParams.ExactMatch = utils.BoolToPtr(true)         // 精准匹配
	response, err := c.client.ListDomains(requestParams)
	if err != nil {
		return HuaweiCdnQueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询加速域名", requestParams, response)
	}

	for _, item := range *response.Domains {
		if strings.EqualFold(utils.TrimDots(utils.PtrToString(item.DomainName)), utils.TrimDots(fullDomain)) {
			sources := *item.Sources
			return HuaweiCdnQueryDomainResponse{
				RecordID:    utils.PtrToString(item.Id),
				FullDomain:  utils.TrimDots(utils.PtrToString(item.DomainName)),
				RecordType:  sources[0].OriginType.Value(),
				OriginValue: sources[0].IpOrDomain,
			}, nil
		}
	}

	return HuaweiCdnQueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建加速域名
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/debug?version=v2&api=CreateDomain
// fullDomain: 完整域名
// originValue: 源站地址
func (c *HuaweiCdnClient) CreateDomain(ctx context.Context, fullDomain string, originValue string) error {

	// 请求
	requestParams := &model.CreateDomainRequest{}
	requestParams.Body = &model.CreateDomainRequestBody{
		Domain: &model.DomainBody{
			DomainName:   fullDomain,
			BusinessType: model.GetDomainBodyBusinessTypeEnum().WEB,
			Sources: []model.SourcesRequestBody{
				{
					IpOrDomain:    originValue,
					OriginType:    model.GetSourcesRequestBodyOriginTypeEnum().IPADDR,
					ActiveStandby: int32(1),
				},
			},
		},
	}
	response, err := c.client.CreateDomain(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建加速域名", requestParams, response)
	}

	return nil
}

// 更新加速域名
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/sdk?version=v2&api=UpdateDomainFullConfig
// fullDomain: 完整域名
// originValue: 源站地址
func (c *HuaweiCdnClient) UpdateDomain(ctx context.Context, fullDomain string, originValue string) error {

	// 请求
	requestParams := &model.UpdateDomainFullConfigRequest{}
	requestParams.DomainName = fullDomain // 加速域名
	requestParams.Body = &model.ModifyDomainConfigRequestBody{
		Configs: &model.Configs{
			Sources: &[]model.SourcesConfig{
				{
					OriginType: model.GetSourcesRequestBodyOriginTypeEnum().IPADDR.Value(),
					OriginAddr: originValue,
				},
			},
		},
	}
	response, err := c.client.UpdateDomainFullConfig(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新加速域名", requestParams, response)
	}

	return nil
}
