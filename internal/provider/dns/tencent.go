package dns

import (
	"context"
	"fmt"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// TencentDNSConfig 腾讯云DNS配置
type TencentDNSConfig struct {
	Endpoint   string `json:"endpoint,omitempty"` // 域名地址
	SecretID   string `json:"secret_id"`          // 密钥ID
	SecretKey  string `json:"secret_key"`         // 密钥Key
	DomainName string `json:"domain_name"`        // 主域名
}

// Validate 验证腾讯云DNS配置
func (cfg *TencentDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("域名地址不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("密钥ID不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("密钥Key不能为空")
	}
	return nil
}

// NewDefaultTencentDNSConfig 创建默认的腾讯云DNS配置
func NewDefaultTencentDNSConfig() *TencentDNSConfig {
	return &TencentDNSConfig{
		Endpoint: "dnspod.tencentcloudapi.com",
	}
}

// TencentDNSProvider 腾讯云DNS提供程序
type TencentDNSProvider struct {
	config *TencentDNSConfig
	opts   *dnsOptions
	client *dnspod.Client
}

// NewTencentDNSProvider 创建腾讯云DNS提供程序
func NewTencentDNSProvider(ctx context.Context, cfg *TencentDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {

	var err error

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &TencentDNSProvider{
		config: cfg,
		opts:   &dnsOptions{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	credential := common.NewCredential(
		p.config.SecretID,
		p.config.SecretKey,
	)

	// 实例化client选项
	clientConfig := profile.NewClientProfile()
	clientConfig.HttpProfile.Endpoint = p.config.Endpoint

	// 实例化client对象
	p.client, err = dnspod.NewClient(credential, "", clientConfig)
	if err != nil {
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.client.WithHttpTransport(p.opts.httpTransport)
	}

	return p, nil
}

// Check 检查子域名是否存在
func (p *TencentDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error) {

	// 请求参数
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)       // 域名
	request.Subdomain = common.StringPtr(subdomain) // 解析记录的主机头
	if recordType != "" {
		request.RecordType = common.StringPtr(recordType) // 获取某种类型的解析记录
	}
	// 请求
	response, err := p.client.DescribeRecordListWithContext(ctx, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Response.RecordList {
		if strings.EqualFold(utils.PtrToString(item.Name), subdomain) {
			if recordType != "" && !strings.EqualFold(utils.PtrToString(item.Type), recordType) {
				continue
			}
			recordID = cast.ToString(*item.RecordId)
			value = utils.PtrToString(item.Value)
			return recordID, value, nil
		}
	}

	return "", "", fmt.Errorf("未找到记录")
}

// Update 更新 DNS 记录
func (p *TencentDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	if recordID == "" {
		// 请求参数
		request := dnspod.NewCreateRecordRequest()
		request.Domain = common.StringPtr(domain)         // 域名
		request.RecordType = common.StringPtr(recordType) // 记录类型
		request.RecordLine = common.StringPtr("默认")       // 记录线路
		request.Value = common.StringPtr(value)           // 记录值
		request.SubDomain = common.StringPtr(subdomain)   // 主机记录
		// 请求
		response, err := p.client.CreateRecordWithContext(ctx, request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := dnspod.NewModifyRecordRequest()
		request.Domain = common.StringPtr(domain)                    // 域名
		request.RecordType = common.StringPtr(recordType)            // 记录类型
		request.RecordLine = common.StringPtr("默认")                  // 记录线路
		request.Value = common.StringPtr(value)                      // 记录值
		request.RecordId = common.Uint64Ptr(cast.ToUint64(recordID)) // 记录 ID
		request.SubDomain = common.StringPtr(subdomain)              // 主机记录
		// 请求
		response, err := p.client.ModifyRecordWithContext(ctx, request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *TencentDNSProvider) Save(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *TencentDNSProvider) Name() string {
	return "腾讯云 DNS"
}
