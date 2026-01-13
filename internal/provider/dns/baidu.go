package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"vibe_ip/internal/provider"

	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/baidubce/bce-sdk-go/util"
)

// 百度云 DNS 自定义配置
type BaiduDNSCustomConfig struct {
}

// BaiduDNSConfig 百度云DNS配置
type BaiduDNSConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 域名
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	SecretAccessKey string `json:"secret_access_key"`  // 密钥Key
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证百度云DNS配置
func (cfg *BaiduDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("域名不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("密钥ID不能为空")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("密钥Key不能为空")
	}
	return nil
}

// NewDefaultBaiduDNSConfig 创建默认的百度云DNS配置
func NewDefaultBaiduDNSConfig() *BaiduDNSConfig {
	return &BaiduDNSConfig{
		Endpoint: "dns.baidubce.com",
	}
}

// BaiduDNSProvider 百度云DNS提供程序
type BaiduDNSProvider struct {
	config *BaiduDNSConfig
	opts   *dnsOptions
	client *dns.Client
}

// NewBaiduDNSProvider 创建百度云DNS提供程序
func NewBaiduDNSProvider(ctx context.Context, cfg *BaiduDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {

	var err error

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &BaiduDNSProvider{
		config: cfg,
		opts:   &dnsOptions{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化client对象
	p.client, err = dns.NewClient(p.config.AccessKeyID, p.config.SecretAccessKey, p.config.Endpoint)
	if err != nil {
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.client.HTTPClient = &http.Client{
			Transport: p.opts.httpTransport,
		}
	}

	return p, nil
}

// Check 检查子域名是否存在
func (p *BaiduDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error) {

	// 请求参数
	request := &dns.ListRecordRequest{}
	request.Rr = subdomain // 主机记录
	request.MaxKeys = 1000 // 每页包含的最大数量，最大数量通常不超过1000，缺省值为1000
	// 请求
	response, err := p.client.ListRecord(domain, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Records {
		if strings.EqualFold(item.Rr, subdomain) {
			if recordType != "" && !strings.EqualFold(item.Type, recordType) {
				continue
			}
			recordID = item.Id
			value = item.Value
			return recordID, value, nil
		}
	}

	return "", "", fmt.Errorf("未找到记录")
}

// Update 更新 DNS 记录
func (p *BaiduDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	if recordID == "" {
		// 请求参数
		request := &dns.CreateRecordRequest{}
		request.Rr = subdomain    // 主机记录
		request.Type = recordType // 解析记录类型
		request.Value = value     // 记录值
		// 请求
		err = p.client.CreateRecord(domain, request, util.NewUUID())
		printJSON("创建域名记录", request, nil)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := &dns.UpdateRecordRequest{}
		request.Rr = subdomain    // 主机记录
		request.Type = recordType // 解析记录类型
		request.Value = value     // 记录值
		// 请求
		err = p.client.UpdateRecord(domain, recordID, request, util.NewUUID())
		printJSON("更新域名记录", request, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *BaiduDNSProvider) Save(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *BaiduDNSProvider) Name() string {
	return "百度云 DNS"
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *BaiduDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	return nil
}
