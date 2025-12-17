package dns_helper

import (
	"context"
	"fmt"
	"net/http"
	"vibe_ip/internal/utils"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/baidubce/bce-sdk-go/util"
	"go.dtapp.net/library/contrib/http_log"
)

type BaiduDnsClient struct {
	config struct {
		AccessKeyID     string // 密钥ID
		SecretAccessKey string // 密钥密钥
	}

	client *dns.Client
}

func NewBaiduDnsClient(ctx context.Context, accessKeyID string, secretAccessKey string) (*BaiduDnsClient, error) {

	var err error

	c := &BaiduDnsClient{}
	c.config.AccessKeyID = accessKeyID
	c.config.SecretAccessKey = secretAccessKey

	c.client, err = dns.NewClient(c.config.AccessKeyID, c.config.SecretAccessKey, "dns.baidubce.com")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
// func (c *BaiduDnsClient) SetDebug(debug bool) *BaiduDnsClient {
// 	c.client.WithDebug(debug)
// 	return c
// }

// 设置HTTP日志记录器
func (c *BaiduDnsClient) SetLog(log *http_log.LoggingTransport) *BaiduDnsClient {
	if log == nil {
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
	c.client.HTTPClient.Transport = log
	return c
}

// 获取解析记录列表
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#查询解析记录列表
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *BaiduDnsClient) QueryDomain(ctx context.Context, domainName string, subDomain string, originValue string) (QueryDomainResponse, error) {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return QueryDomainResponse{}, err
	}

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

	for _, item := range response.Records {
		if item.Rr == subDomain {
			if recordType != "" && item.Type != recordType {
				continue
			}
			return QueryDomainResponse{
				DomainID:    item.Id,
				DomainName:  item.Rr,
				OriginType:  item.Type,
				OriginValue: item.Value,
			}, nil
		}
	}

	return QueryDomainResponse{}, fmt.Errorf("未找到记录")
}

// 添加记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#添加解析记录
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *BaiduDnsClient) CreateDomain(ctx context.Context, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &dns.CreateRecordRequest{}
	requestParams.Rr = subDomain      // 主机记录
	requestParams.Type = recordType   // 解析记录类型
	requestParams.Value = originValue // 记录值
	err = c.client.CreateRecord(domainName, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}

// 更新域名记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#修改解析记录
// domainID: 域名ID
// domainName: 主域名
// subDomain: 子域名
// originValue: 源站地址
func (c *BaiduDnsClient) UpdateDomain(ctx context.Context, domainID string, domainName string, subDomain string, originValue string) error {

	// 提取域名中的主域名和子域，同时判断记录类型
	_, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, originValue)
	if err != nil {
		return err
	}

	// 请求
	requestParams := &dns.UpdateRecordRequest{}
	requestParams.Rr = subDomain      // 主机记录
	requestParams.Type = recordType   // 解析记录类型
	requestParams.Value = originValue // 记录值
	err = c.client.UpdateRecord(domainName, domainID, requestParams, util.NewUUID())
	if _, ok := err.(*bce.BceClientError); ok {
		return fmt.Errorf("客户端出现错误: %w", err)
	}
	if _, ok := err.(*bce.BceServiceError); ok {
		return fmt.Errorf("服务出现错误: %w", err)
	}
	if err != nil {
		return err
	}

	return nil
}
