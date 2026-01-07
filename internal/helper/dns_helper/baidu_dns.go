package dns_helper

import (
	"context"
	"fmt"
	"net/http"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/baidubce/bce-sdk-go/util"
	"go.dtapp.net/library/contrib/http_log"
)

type BaiduDnsConfig struct {
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥Key
	DomainName      string `json:"domain_name"`       // 主域名
}

type BaiduDnsClient struct {
	config *BaiduDnsConfig // 配置
	debug  bool            // 是否开启调试模式
	client *dns.Client
}

// 百度云 智能云解析DNS
// https://cloud.baidu.com/product/dns.html
// https://cloud.baidu.com/doc/DNS/s/fm4tmlvcs
func NewBaiduDnsClient(ctx context.Context, config *BaiduDnsConfig) (*BaiduDnsClient, error) {

	var err error

	c := &BaiduDnsClient{}
	c.config = config

	c.client, err = dns.NewClient(c.config.AccessKeyID, c.config.SecretAccessKey, "dns.baidubce.com")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *BaiduDnsClient) SetDebug(debug bool) *BaiduDnsClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *BaiduDnsClient) SetLog(loggingTransport *http_log.LoggingTransport) *BaiduDnsClient {
	if loggingTransport == nil {
		return c
	}
	if c.client.HTTPClient == nil {
		c.client.HTTPClient = &http.Client{}
	}
	// 检查 c.client 是否初始化
	if c.client == nil || c.client.HTTPClient == nil {
		return c
	}
	// 将拦截器本身赋值给 Transport
	c.client.HTTPClient.Transport = loggingTransport
	return c
}

// 查询域名记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#查询解析记录列表
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
func (c *BaiduDnsClient) QueryDomain(ctx context.Context, domainName string, subDomain string, recordType string) (QueryDomainResponse, error) {

	// 请求
	requestParams := &dns.ListRecordRequest{}
	requestParams.Rr = subDomain // 主机记录
	requestParams.MaxKeys = 1000 // 每页包含的最大数量，最大数量通常不超过1000，缺省值为1000
	response, err := c.client.ListRecord(domainName, requestParams)
	if _, ok := err.(*bce.BceClientError); ok {
		return QueryDomainResponse{}, fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return QueryDomainResponse{}, fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return QueryDomainResponse{}, err
	}
	if c.debug {
		printJSON("查询域名记录", requestParams, response)
	}

	for _, item := range response.Records {
		if item.Rr == subDomain {
			if recordType != "" && item.Type != recordType {
				continue
			}
			return QueryDomainResponse{
				RecordID:    item.Id,
				FullDomain:  item.Rr,
				RecordType:  item.Type,
				OriginValue: item.Value,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 创建域名记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#添加解析记录
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
// originValue: 源站地址
func (c *BaiduDnsClient) CreateDomain(ctx context.Context, domainName string, subDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &dns.CreateRecordRequest{}
	requestParams.Rr = subDomain      // 主机记录
	requestParams.Type = recordType   // 解析记录类型
	requestParams.Value = originValue // 记录值
	err := c.client.CreateRecord(domainName, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("创建域名记录", requestParams, nil)
	}

	return nil
}

// 更新域名记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#修改解析记录
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// recordType: 记录类型
// originValue: 源站地址
func (c *BaiduDnsClient) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, recordType string, originValue string) error {

	// 请求
	requestParams := &dns.UpdateRecordRequest{}
	requestParams.Rr = subDomain      // 主机记录
	requestParams.Type = recordType   // 解析记录类型
	requestParams.Value = originValue // 记录值
	err := c.client.UpdateRecord(domainName, domainID, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}
	if c.debug {
		printJSON("更新域名记录", requestParams, nil)
	}

	return nil
}
