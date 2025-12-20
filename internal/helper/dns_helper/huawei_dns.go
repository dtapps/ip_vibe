package dns_helper

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
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	regionDns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
	"go.dtapp.net/library/contrib/http_log"
)

type HuaweiDnsConfig struct {
	region          string // 区域
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥密钥
	ZoneID          string `json:"zone_id"`           // 域名ID
	DomainName      string `json:"domain_name"`       // 主域名
}

type HuaweiDnsClient struct {
	config  *HuaweiDnsConfig // 配置
	debug   bool             // 是否开启调试模式
	auth    *auth.BasicCredentials
	region  *region.Region
	builder *core.HcHttpClient
	client  *dns.DnsClient
}

// 华为云 DNS
// https://support.huaweicloud.com/dns/index.html
func NewHuaweiDnsClient(ctx context.Context, config *HuaweiDnsConfig) (*HuaweiDnsClient, error) {

	var err error

	c := &HuaweiDnsClient{}
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
	c.region, err = regionDns.SafeValueOf(c.config.region)
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	c.builder, err = dns.DnsClientBuilder().
		WithRegion(c.region).
		WithCredential(c.auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}
	c.client = dns.NewDnsClient(c.builder)

	return c, nil
}

// 设置调试模式
func (c *HuaweiDnsClient) SetDebug(debug bool) *HuaweiDnsClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *HuaweiDnsClient) SetLog(loggingTransport *http_log.LoggingTransport) *HuaweiDnsClient {
	if loggingTransport == nil {
		return c
	}
	// 自定义HTTP传输
	httpConfig := config.DefaultHttpConfig().WithHttpRoundTripper(loggingTransport)
	// 实例化client对象
	builder, err := dns.DnsClientBuilder().
		WithRegion(c.region).
		WithCredential(c.auth).
		WithHttpConfig(httpConfig).
		SafeBuild()
	if err != nil {
		return c
	}
	c.builder = builder
	c.client = dns.NewDnsClient(c.builder)
	return c
}

// 查询域名记录
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=ShowRecordSetByZone
// fullDomain: 完整域名
// recordType: 记录类型
func (c *HuaweiDnsClient) QueryDomain(ctx context.Context, fullDomain string, recordType string) (QueryDomainResponse, error) {

	// 请求
	requestParams := &model.ShowRecordSetByZoneRequest{}
	requestParams.Limit = utils.Int32ToPtr(500) // 分页查询时配置每页返回的资源个数
	requestParams.Offset = utils.Int32ToPtr(0)  // 分页查询起始偏移量,表示从偏移量的下一个资源开始查询
	if recordType != "" {
		requestParams.Type = utils.StringToPtr(recordType) // 记录集的类型
	}
	requestParams.Name = utils.StringToPtr(fullDomain)    // 待查询的记录集的域名中包含此name
	requestParams.SearchMode = utils.StringToPtr("equal") // 查询条件搜索模式
	requestParams.ZoneId = c.config.ZoneID                // 域名ID
	response, err := c.client.ShowRecordSetByZone(requestParams)
	if err != nil {
		return QueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询域名记录", requestParams, response)
	}

	for _, item := range *response.Recordsets {
		if utils.PtrToString(item.Name) == fullDomain {
			if recordType != "" && utils.PtrToString(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    utils.PtrToString(item.Id),
				FullDomain:  utils.PtrToString(item.Name),
				RecordType:  utils.PtrToString(item.Type),
				OriginValue: strings.Join(utils.PtrToStrings(item.Records), ","),
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=CreateRecordSetWithLine
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *HuaweiDnsClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &model.CreateRecordSetWithLineRequest{}
	requestParams.ZoneId = c.config.ZoneID // 域名ID
	requestParams.Body = &model.CreateRecordSetWithLineRequestBody{
		Records: utils.StringsToPtr([]string{originValue}), // 解析记录的值
		Type:    recordType,                                // 记录集的类型
		Name:    fullDomain,                                // 域名
	}
	response, err := c.client.CreateRecordSetWithLine(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建域名记录", requestParams, response)
	}

	return nil
}

// 更新域名记录
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/sdk?api=UpdateRecordSets
// domainID: 域名ID
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *HuaweiDnsClient) UpdateDomain(ctx context.Context, domainID string, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &model.UpdateRecordSetsRequest{}
	requestParams.ZoneId = c.config.ZoneID // 域名ID
	requestParams.RecordsetId = domainID   // 记录集ID
	requestParams.Body = &model.UpdateRecordSetsReq{
		Records: utils.StringsToPtr([]string{originValue}), // 解析记录的值
		Type:    recordType,                                // 记录集的类型
		Name:    fullDomain,                                // 域名
	}
	response, err := c.client.UpdateRecordSets(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新域名记录", requestParams, response)
	}

	return nil
}
