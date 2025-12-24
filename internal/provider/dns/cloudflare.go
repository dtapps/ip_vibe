package dns

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

// CloudfareDNSConfig CloudfareDNS配置
type CloudfareDNSConfig struct {
	ApiKey     string `json:"api_key"`     // API密钥
	ZoneID     string `json:"zone_id"`     // 区域ID
	DomainName string `json:"domain_name"` // 主域名
}

// Validate 验证CloudfareDNS配置
func (config *CloudfareDNSConfig) Validate() error {
	if config.ApiKey == "" {
		return fmt.Errorf("API密钥不能为空")
	}
	if config.ZoneID == "" {
		return fmt.Errorf("区域ID不能为空")
	}
	return nil
}

// NewDefaultCloudfareDNSConfig 创建默认的CloudfareDNS配置
func NewDefaultCloudfareDNSConfig() *CloudfareDNSConfig {
	return &CloudfareDNSConfig{}
}

type CloudfareDNSProvider struct {
	config *CloudfareDNSConfig
	opts   *dnsOptions
	client *cloudflare.Client
}

func NewCloudfareDNSProvider(ctx context.Context, config *CloudfareDNSConfig, opts ...DnsOption) (p *CloudfareDNSProvider, err error) {

	if config == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p = &CloudfareDNSProvider{
		config: config,
		opts:   &dnsOptions{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化client选项
	clientOpts := make([]option.RequestOption, 0)
	clientOpts = append(clientOpts, option.WithAPIKey(p.config.ApiKey))

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		clientOpts = append(clientOpts, option.WithHTTPClient(&http.Client{
			Transport: p.opts.httpTransport,
		}))
	}

	// 实例化client对象
	p.client = cloudflare.NewClient(clientOpts...)

	return p, nil
}

// Check 检查子域名是否存在
func (p *CloudfareDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (id string, ip string, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := dns.RecordListParams{}
	request.ZoneID = cloudflare.String(p.config.ZoneID)
	request.Name = cloudflare.F(dns.RecordListParamsName{Exact: cloudflare.String(fullDomain)})
	if recordType != "" {
		request.Type = cloudflare.F(dns.RecordListParamsType(recordType))
	}
	// 请求
	response, err := p.client.DNS.Records.List(ctx, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Result {
		if item.Name == fullDomain {
			if recordType != "" && item.Type != dns.RecordResponseType(recordType) {
				continue
			}
			id = item.ID
			ip = item.Content
			return id, ip, nil
		}
	}

	return "", "", fmt.Errorf("未找到记录")
}

// Update 更新 DNS 记录
func (p *CloudfareDNSProvider) Update(ctx context.Context, id string, domain string, subdomain string, ip string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if id == "" {
		// 请求参数
		requestBody := dns.ARecordParam{}
		requestBody.Name = cloudflare.String(fullDomain)
		requestBody.Type = cloudflare.F(dns.ARecordType(recordType))
		requestBody.Content = cloudflare.String(ip)
		request := dns.RecordNewParams{
			ZoneID: cloudflare.String(p.config.ZoneID),
			Body:   requestBody,
		}
		// 请求
		response, err := p.client.DNS.Records.New(ctx, request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if id != "" {
		// 请求参数
		requestBody := dns.ARecordParam{}
		requestBody.Name = cloudflare.String(fullDomain)
		requestBody.Type = cloudflare.F(dns.ARecordType(recordType))
		requestBody.Content = cloudflare.String(ip)
		request := dns.RecordEditParams{
			ZoneID: cloudflare.String(p.config.ZoneID),
			Body:   requestBody,
		}
		// 请求
		response, err := p.client.DNS.Records.Edit(ctx, id, request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *CloudfareDNSProvider) Save(ctx context.Context, id, domain, subdomain, ip, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *CloudfareDNSProvider) Name() string {
	return "阿里云 DNS"
}
