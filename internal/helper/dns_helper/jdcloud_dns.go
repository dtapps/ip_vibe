package dns_helper

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/models"
	"github.com/spf13/cast"
	"go.dtapp.net/library/contrib/http_log"
)

type JdcloudDnsConfig struct {
	region     string // 区域
	AccessKey  string `json:"access_key"`  // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥密钥
	ZoneID     string `json:"zone_id"`     // 域名ID
	DomainName string `json:"domain_name"` // 主域名
}

type JdcloudDnsClient struct {
	config *JdcloudDnsConfig // 配置
	debug  bool              // 是否开启调试模式
	auth   *core.Credential
	client *client.DomainserviceClient
}

// TrafficRoute DNS套件
// https://www.volcengine.com/product/trafficroute
func NewJdcloudDnsClient(ctx context.Context, config *JdcloudDnsConfig) (*JdcloudDnsClient, error) {

	c := &JdcloudDnsClient{}
	c.config = config
	c.config.region = "cn-north-1"

	// 认证
	c.auth = core.NewCredentials(c.config.AccessKey, c.config.SecretKey)

	// 实例化client对象
	c.client = client.NewDomainserviceClient(c.auth)

	return c, nil
}

// 设置调试模式
func (c *JdcloudDnsClient) SetDebug(debug bool) *JdcloudDnsClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *JdcloudDnsClient) SetLog(loggingTransport *http_log.LoggingTransport) *JdcloudDnsClient {
	if loggingTransport == nil {
		return c
	}
	return c
}

// 查询域名记录
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/describeresourcerecord
// fullDomain: 完整域名
// recordType: 记录类型
func (c *JdcloudDnsClient) QueryDomain(ctx context.Context, fullDomain string, recordType string) (QueryDomainResponse, error) {

	// 请求
	request := apis.NewDescribeResourceRecordRequestWithAllParams(
		c.config.region,
		c.config.ZoneID,
		utils.IntToPtr(1),
		utils.IntToPtr(500),
		utils.StringToPtr(fullDomain),
	)
	response, err := c.client.DescribeResourceRecord(request)
	if err != nil {
		return QueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询域名记录", request, response)
	}

	for _, item := range response.Result.DataList {
		if strings.EqualFold(utils.TrimDots(item.HostValue), utils.TrimDots(fullDomain)) {
			if recordType != "" && item.Type != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    cast.ToString(item.Id),
				FullDomain:  utils.TrimDots(item.HostValue),
				RecordType:  item.Type,
				OriginValue: item.HostValue,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/createresourcerecord
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *JdcloudDnsClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &models.AddRR{}
	requestParams.HostRecord = fullDomain // 主机记录
	requestParams.HostValue = originValue // 解析记录的值
	requestParams.Type = recordType       // 解析的类型
	request := apis.NewCreateResourceRecordRequest(c.config.region, c.config.ZoneID, requestParams)
	response, err := c.client.CreateResourceRecord(request)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建域名记录", requestParams, response)
	}

	return nil
}

// 更新域名记录
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/modifyresourcerecord
// domainID: 域名ID
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *JdcloudDnsClient) UpdateDomain(ctx context.Context, domainID string, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &models.UpdateRR{}
	requestParams.HostRecord = fullDomain // 主机记录
	requestParams.HostValue = originValue // 解析记录的值
	requestParams.Type = recordType       // 解析的类型
	request := apis.NewModifyResourceRecordRequest(c.config.region, c.config.ZoneID, domainID, requestParams)
	response, err := c.client.ModifyResourceRecord(request)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新域名记录", requestParams, response)
	}

	return nil
}
