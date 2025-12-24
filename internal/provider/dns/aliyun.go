package dns

import (
	"context"
	"fmt"
	"vibe_ip/internal/helper/aliyun_helper"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
)

// AliyunDNSConfig 阿里云DNS配置
type AliyunDNSConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 域名地址
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	AccessKeySecret string `json:"access_key_secret"`  // 密钥Key
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证阿里云DNS配置
func (cfg *AliyunDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("域名地址不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("密钥ID不能为空")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("密钥Key不能为空")
	}
	return nil
}

// NewDefaultAliyunDNSConfig 创建默认的阿里云DNS配置
func NewDefaultAliyunDNSConfig() *AliyunDNSConfig {
	return &AliyunDNSConfig{
		Endpoint: "alidns.aliyuncs.com",
	}
}

type AliyunDNSProvider struct {
	config *AliyunDNSConfig
	opts   *dnsOptions
	client *alidns20150109.Client
}

func NewAliyunDNSProvider(ctx context.Context, cfg *AliyunDNSConfig, opts ...DnsOption) (p *AliyunDNSProvider, err error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p = &AliyunDNSProvider{
		config: cfg,
		opts:   &dnsOptions{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	credential, err := credential.NewCredential(&credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(p.config.AccessKeyID),
		AccessKeySecret: tea.String(p.config.AccessKeySecret),
	})
	if err != nil {
		return nil, err
	}

	// 实例化client选项
	clientConfig := &openapi.Config{
		Credential: credential,
		Endpoint:   tea.String(p.config.Endpoint),
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		httpHandler := aliyun_helper.NewHttpHandler(p.opts.httpTransport)
		clientConfig.HttpClient = httpHandler
	}

	// 实例化client对象
	p.client, err = alidns20150109.NewClient(clientConfig)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Check 检查子域名是否存在
func (p *AliyunDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (id string, ip string, err error) {

	// 请求参数
	request := &alidns20150109.DescribeDomainRecordsRequest{}
	request.SearchMode = tea.String("COMBINATION") // 搜索模式
	request.DomainName = tea.String(domain)        // 域名名称
	request.RRKeyWord = tea.String(subdomain)      // 主机记录的关键字
	if recordType != "" {
		request.TypeKeyWord = tea.String(recordType) // 解析类型的关键字
	}
	// 请求
	response, err := p.client.DescribeDomainRecordsWithContext(ctx, request, nil)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Body.DomainRecords.Record {
		if tea.StringValue(item.RR) == subdomain && tea.StringValue(item.DomainName) == domain {
			if recordType != "" && tea.StringValue(item.Type) != recordType {
				continue
			}
			id = tea.StringValue(item.RecordId)
			ip = tea.StringValue(item.Value)
			return id, ip, nil
		}
	}

	return "", "", fmt.Errorf("未找到记录")
}

// Update 更新 DNS 记录
func (p *AliyunDNSProvider) Update(ctx context.Context, id string, domain string, subdomain string, ip string, recordType string) (err error) {

	if id == "" {
		// 请求参数
		request := &alidns20150109.AddDomainRecordRequest{}
		request.DomainName = tea.String(domain) // 域名名称
		request.RR = tea.String(subdomain)      // 主机记录
		request.Type = tea.String(recordType)   // 解析记录类型
		request.Value = tea.String(ip)          // 记录值
		// 请求
		response, err := p.client.AddDomainRecordWithContext(ctx, request, nil)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if id != "" {
		// 请求参数
		request := &alidns20150109.UpdateDomainRecordRequest{}
		request.RecordId = tea.String(id)     // 解析记录的ID
		request.RR = tea.String(subdomain)    // 主机记录
		request.Type = tea.String(recordType) // 解析记录类型
		request.Value = tea.String(ip)        // 记录值
		// 请求
		response, err := p.client.UpdateDomainRecordWithContext(ctx, request, nil)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *AliyunDNSProvider) Save(ctx context.Context, id, domain, subdomain, ip, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *AliyunDNSProvider) Name() string {
	return "阿里云 DNS"
}
