package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	dns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/model"
	regionDns "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/dns/v2/region"
)

// HuaweiDNSConfig 华为云 DNS 配置
type HuaweiDNSConfig struct {
	Region          string `json:"region,omitempty"`  // 区域
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥密钥
	ZoneID          string `json:"zone_id"`           // 域名ID
	DomainName      string `json:"domain_name"`       // 主域名
}

// Validate 验证华为云 DNS 配置
func (cfg *HuaweiDNSConfig) Validate() error {
	if cfg.Region == "" {
		return fmt.Errorf("Region 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("SecretAccessKey 不能为空")
	}
	if cfg.ZoneID == "" {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultHuaweiDNSConfig 创建默认的华为云 DNS 配置
func NewDefaultHuaweiDNSConfig() *HuaweiDNSConfig {
	return &HuaweiDNSConfig{
		Region: "cn-south-1",
	}
}

// ParseCustomConfig 解析 华为云 DNS 自定义配置
func (p *HuaweiDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// HuaweiDNSProvider 华为云 DNS 提供程序
type HuaweiDNSProvider struct {
	config *HuaweiDNSConfig
	opts   *dnsOptions
	client *dns.DnsClient
}

// NewHuaweiDNSProvider 创建华为云 DNS 提供程序
func NewHuaweiDNSProvider(ctx context.Context, cfg *HuaweiDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {

	var err error

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &HuaweiDNSProvider{
		config: cfg,
		opts:   &dnsOptions{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 认证
	auth, err := basic.NewCredentialsBuilder().
		WithAk(p.config.AccessKeyID).
		WithSk(p.config.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	// 区域
	region, err := regionDns.SafeValueOf(p.config.Region)
	if err != nil {
		return nil, err
	}

	clientBuilder := dns.DnsClientBuilder().
		WithRegion(region).
		WithCredential(auth)

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		clientBuilder.WithHttpConfig(config.DefaultHttpConfig().WithHttpRoundTripper(p.opts.httpTransport))
	}

	// 实例化client对象
	builder, err := clientBuilder.SafeBuild()
	if err != nil {
		return nil, err
	}
	p.client = dns.NewDnsClient(builder)

	return p, nil
}

// Check 检查子域名是否存在
func (p *HuaweiDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := &model.ShowRecordSetByZoneRequest{}
	request.Limit = utils.Int32ToPtr(500) // 分页查询时配置每页返回的资源个数
	request.Offset = utils.Int32ToPtr(0)  // 分页查询起始偏移量,表示从偏移量的下一个资源开始查询
	if recordType != "" {
		request.Type = utils.StringToPtr(recordType) // 记录集的类型
	}
	request.Name = utils.StringToPtr(fullDomain)    // 待查询的记录集的域名中包含此name
	request.SearchMode = utils.StringToPtr("equal") // 查询条件搜索模式
	request.ZoneId = p.config.ZoneID                // 域名ID
	// 请求
	response, err := p.client.ShowRecordSetByZone(request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range *response.Recordsets {
		if strings.EqualFold(utils.TrimDots(utils.PtrToString(item.Name)), utils.TrimDots(fullDomain)) {
			if recordType != "" && utils.PtrToString(item.Type) != recordType {
				continue
			}
			recordID = utils.PtrToString(item.Id)
			value = strings.Join(utils.PtrToStrings(item.Records), ",")
			return recordID, value, nil
		}
	}

	return "", "", fmt.Errorf("未找到记录")
}

// Update 更新 DNS 记录
func (p *HuaweiDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if recordID == "" {
		// 请求参数
		request := &model.CreateRecordSetWithLineRequest{}
		request.ZoneId = p.config.ZoneID // 域名ID
		request.Body = &model.CreateRecordSetWithLineRequestBody{
			Records: utils.StringsToPtr([]string{value}), // 解析记录的值
			Type:    recordType,                          // 记录集的类型
			Name:    fullDomain,                          // 域名
		}
		// 请求
		response, err := p.client.CreateRecordSetWithLine(request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := &model.UpdateRecordSetsRequest{}
		request.ZoneId = p.config.ZoneID // 域名ID
		request.RecordsetId = recordID   // 记录集ID
		request.Body = &model.UpdateRecordSetsReq{
			Records: utils.StringsToPtr([]string{value}), // 解析记录的值
			Type:    recordType,                          // 记录集的类型
			Name:    fullDomain,                          // 域名
		}
		// 请求
		response, err := p.client.UpdateRecordSets(request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *HuaweiDNSProvider) Save(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *HuaweiDNSProvider) Name() string {
	return "华为云 DNS"
}

// 华为云 DNS 自定义配置
type HuaweiDNSCustomConfig struct {
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *HuaweiDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	return nil
}
