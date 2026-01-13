package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
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
	cfg  *HuaweiDNSConfig
	opts *dnsOptions

	CustomConfig *HuaweiDNSCustomConfig
	Provider     struct {
		Auth          *auth.BasicCredentials
		Region        *region.Region
		ClientBuilder *core.HcHttpClientBuilder
		Builder       *core.HcHttpClient
		Client        *dns.DnsClient
	}
}

// NewHuaweiDNSProvider 创建华为云 DNS 提供程序
func NewHuaweiDNSProvider(ctx context.Context, cfg *HuaweiDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &HuaweiDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		CustomConfig: &HuaweiDNSCustomConfig{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 认证
	p.Provider.Auth, err = basic.NewCredentialsBuilder().
		WithAk(p.cfg.AccessKeyID).
		WithSk(p.cfg.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	// 区域
	p.Provider.Region, err = regionDns.SafeValueOf(p.cfg.Region)
	if err != nil {
		return nil, err
	}

	p.Provider.ClientBuilder = dns.DnsClientBuilder().
		WithRegion(p.Provider.Region).
		WithCredential(p.Provider.Auth)

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.ClientBuilder.WithHttpConfig(config.DefaultHttpConfig().
			WithHttpRoundTripper(p.opts.httpTransport.Instance(nil)))
	}

	p.Provider.Builder, err = p.Provider.ClientBuilder.SafeBuild()
	if err != nil {
		return nil, err
	}

	// 实例化client对象
	p.Provider.Client = dns.NewDnsClient(p.Provider.Builder)

	return p, nil
}

// Check 检查子域名是否存在
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=ShowRecordSetByZone
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
	request.ZoneId = p.cfg.ZoneID                   // 域名ID
	// 请求
	response, err := p.Provider.Client.ShowRecordSetByZone(request)
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

	return "", "", provider.ErrRecordNotFound
}

// Update 更新 DNS 记录
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=CreateRecordSetWithLine
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/sdk?api=UpdateRecordSets
func (p *HuaweiDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if recordID == "" {
		// 请求参数
		request := &model.CreateRecordSetWithLineRequest{}
		request.ZoneId = p.cfg.ZoneID // 域名ID
		request.Body = &model.CreateRecordSetWithLineRequestBody{
			Records:     utils.StringsToPtr([]string{value}),           // 解析记录的值
			Type:        recordType,                                    // 记录集的类型
			Name:        fullDomain,                                    // 域名
			Description: utils.StringToPtr(p.CustomConfig.Description), // 记录集的描述信息
		}
		// 请求
		response, err := p.Provider.Client.CreateRecordSetWithLine(request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := &model.UpdateRecordSetsRequest{}
		request.ZoneId = p.cfg.ZoneID  // 域名ID
		request.RecordsetId = recordID // 记录集ID
		request.Body = &model.UpdateRecordSetsReq{
			Records:     utils.StringsToPtr([]string{value}),           // 解析记录的值
			Type:        recordType,                                    // 记录集的类型
			Name:        fullDomain,                                    // 域名
			Description: utils.StringToPtr(p.CustomConfig.Description), // 记录集的描述信息
		}
		// 请求
		response, err := p.Provider.Client.UpdateRecordSets(request)
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
	Description string `json:"description"` // 记录集的描述信息
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *HuaweiDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
