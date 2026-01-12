package dns_helper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"vibe_ip/internal/utils"

	"github.com/volcengine/volcengine-go-sdk/service/dns"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
	"go.dtapp.net/library/contrib/http_log"
)

type VolcengineDnsConfig struct {
	region     string // 区域
	AccessKey  string `json:"access_key"`  // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥密钥
	ZoneID     int64  `json:"zone_id"`     // 域名ID
	DomainName string `json:"domain_name"` // 主域名
}

type VolcengineDnsClient struct {
	config *VolcengineDnsConfig // 配置
	debug  bool                 // 是否开启调试模式
	auth   *credentials.Credentials
	cfg    *volcengine.Config
	sess   *session.Session
	client *dns.DNS
}

// TrafficRoute DNS套件
// https://www.volcengine.com/product/trafficroute
func NewVolcengineDnsClient(ctx context.Context, config *VolcengineDnsConfig) (*VolcengineDnsClient, error) {

	var err error

	c := &VolcengineDnsClient{}
	c.config = config
	c.config.region = "cn-beijing"

	// 认证
	c.auth = credentials.NewStaticCredentials(c.config.AccessKey, c.config.SecretKey, "")

	// 配置
	c.cfg = volcengine.NewConfig().WithRegion(c.config.region).WithCredentials(c.auth)

	// 会话
	c.sess, err = session.NewSession(c.cfg)
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	c.client = dns.New(c.sess)

	return c, nil
}

// 设置调试模式
func (c *VolcengineDnsClient) SetDebug(debug bool) *VolcengineDnsClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *VolcengineDnsClient) SetLog(loggingTransport *http_log.LoggingRoundTripper) *VolcengineDnsClient {
	if loggingTransport == nil {
		return c
	}
	// 构建自定义的 HTTP Client
	httpClient := &http.Client{
		Transport: loggingTransport,
	}
	// 配置
	cfg := volcengine.NewConfig().WithRegion(c.config.region).WithCredentials(c.auth).WithHTTPClient(httpClient)
	// 会话
	sess, err := session.NewSession(cfg)
	if err != nil {
		return c
	}
	c.cfg = cfg
	c.sess = sess
	// 实例化client对象
	c.client = dns.New(sess)
	return c
}

// 查询域名记录
// https://api.volcengine.com/api-explorer?action=ListRecords
// fullDomain: 完整域名
// recordType: 记录类型
func (c *VolcengineDnsClient) QueryDomain(ctx context.Context, fullDomain string, recordType string) (QueryDomainResponse, error) {

	// 请求
	requestParams := &dns.ListRecordsInput{}
	requestParams.Host = volcengine.String(fullDomain)    // 主机记录
	requestParams.PageNumber = volcengine.Int32(1)        // 页码
	requestParams.PageSize = volcengine.Int32(500)        // 每页包含的解析记录的数量
	requestParams.SearchMode = volcengine.String("exact") // Host 匹配模式
	if recordType != "" {
		requestParams.Type = volcengine.String(recordType) // 记录类型
	}
	requestParams.ZID = volcengine.Int64(c.config.ZoneID) // 域名ID
	response, err := c.client.ListRecords(requestParams)
	if err != nil {
		return QueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询域名记录", requestParams, response)
	}

	for _, item := range response.Records {
		if strings.EqualFold(utils.TrimDots(volcengine.StringValue(item.PQDN)), utils.TrimDots(fullDomain)) {
			if recordType != "" && utils.PtrToString(item.Type) != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    volcengine.StringValue(item.RecordID),
				FullDomain:  utils.TrimDots(volcengine.StringValue(item.PQDN)),
				RecordType:  volcengine.StringValue(item.Type),
				OriginValue: volcengine.StringValue(item.Value),
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://api.volcengine.com/api-explorer?action=CreateRecord
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *VolcengineDnsClient) CreateDomain(ctx context.Context, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &dns.CreateRecordInput{}
	requestParams.Host = volcengine.String(fullDomain)    // 主机记录
	requestParams.Type = volcengine.String(recordType)    // 记录类型
	requestParams.Value = volcengine.String(originValue)  // 解析记录的记录值
	requestParams.ZID = volcengine.Int64(c.config.ZoneID) // 域名ID
	response, err := c.client.CreateRecord(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建域名记录", requestParams, response)
	}

	return nil
}

// 更新域名记录
// https://api.volcengine.com/api-explorer?action=UpdateRecord
// domainID: 域名ID
// fullDomain: 完整域名
// recordType: 记录类型
// originValue: 源站地址
func (c *VolcengineDnsClient) UpdateDomain(ctx context.Context, domainID string, fullDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &dns.UpdateRecordInput{}
	requestParams.Host = volcengine.String(fullDomain)   // 主机记录
	requestParams.Line = volcengine.String("default")    // 解析线路
	requestParams.RecordID = volcengine.String(domainID) // 需要更新的解析记录的 ID
	requestParams.Type = volcengine.String(recordType)   // 记录类型
	requestParams.Value = volcengine.String(originValue) // 解析记录的记录值
	response, err := c.client.UpdateRecord(requestParams)
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新域名记录", requestParams, response)
	}

	return nil
}
