package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/spf13/cast"
)

// CloudfareDNSConfig Cloudfare DNS 配置
type CloudfareDNSConfig struct {
	ApiToken   string `json:"api_token"`   // API令牌
	ZoneID     string `json:"zone_id"`     // 域名ID
	DomainName string `json:"domain_name"` // 主域名
}

// Validate 验证Cloudfare DNS 配置
func (cfg *CloudfareDNSConfig) Validate() error {
	if cfg.ApiToken == "" {
		return fmt.Errorf("ApiToken 不能为空")
	}
	if cfg.ZoneID == "" {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultCloudfareDNSConfig 创建默认的Cloudfare DNS 配置
func NewDefaultCloudfareDNSConfig() *CloudfareDNSConfig {
	return &CloudfareDNSConfig{}
}

// ParseCustomConfig 解析 Cloudfare DNS 自定义配置
func (p *CloudfareDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// CloudfareDNSProvider Cloudfare DNS 提供程序
type CloudfareDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg  *CloudfareDNSConfig
	opts *dnsOptions

	CustomConfig *CloudfareDNSCustomConfig
	Provider     struct {
		Opts   []option.RequestOption
		Client *cloudflare.Client
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData dns.RecordResponse
}

// NewCloudfareDNSProvider 创建Cloudfare DNS 提供程序
func NewCloudfareDNSProvider(ctx context.Context, cfg *CloudfareDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewCloudfareDNSProviderRaw(ctx, cfg, opts...)
}

// NewCloudfareDNSProviderRaw 创建Cloudfare DNS 提供程序
func NewCloudfareDNSProviderRaw(ctx context.Context, cfg *CloudfareDNSConfig, opts ...DnsOption) (*CloudfareDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &CloudfareDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		CustomConfig: &CloudfareDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化client选项
	p.Provider.Opts = append(p.Provider.Opts, option.WithAPIToken(p.cfg.ApiToken))

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Opts = append(p.Provider.Opts, option.WithHTTPClient(&http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}))
	}

	// 实例化
	p.Provider.Client = cloudflare.NewClient(p.Provider.Opts...)

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *CloudfareDNSProvider) Name() string {
	return "Cloudflare DNS"
}

// Check 检查子域名是否存在
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
func (p *CloudfareDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := dns.RecordListParams{}
	request.ZoneID = cloudflare.String(p.cfg.ZoneID)
	request.Name = cloudflare.F(dns.RecordListParamsName{
		Exact: cloudflare.String(fullDomain),
	})
	if recordType != "" {
		request.Type = cloudflare.F(dns.RecordListParamsType(recordType))
	}
	// 请求
	response, err := p.Provider.Client.DNS.Records.List(ctx, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Result == nil || len(response.Result) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range response.Result {

		// 查找域名
		itemFullDomain := item.Name

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = item

			// 查找记录类型
			itemRecordType := cast.ToString(item.Type)
			// 查找记录值
			itemOriginValue := item.Content
			// 查找记录集的描述信息
			itemComment := item.Comment

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = item.ID

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %s，实际 %s", newOriginValue, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录描述
			if !utils.EqualDomainIgnoreCase(itemComment, p.CustomConfig.Comment) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录描述不匹配，期望 %s，实际 %s", p.CustomConfig.Comment, itemComment)
				return p.InternalData, nil
			}

			p.InternalData.IsContinue = false
			return p.InternalData, nil
		}
	}

	// 未找到记录
	p.InternalData.IsContinue = true
	return p.InternalData, nil
}

// Update 更新 DNS 记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/edit/
func (p *CloudfareDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if p.InternalData.RecordID == "" {
		// 请求参数
		requestBody := dns.ARecordParam{}
		requestBody.Name = cloudflare.String(fullDomain)
		requestBody.Type = cloudflare.F(dns.ARecordType(recordType))
		requestBody.Content = cloudflare.String(originValue)
		requestBody.Comment = cloudflare.String(p.CustomConfig.Comment)
		request := dns.RecordNewParams{
			ZoneID: cloudflare.String(p.cfg.ZoneID),
			Body:   requestBody,
		}
		// 请求
		response, err := p.Provider.Client.DNS.Records.New(ctx, request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if p.InternalData.RecordID != "" {
		// 请求参数
		requestBody := dns.ARecordParam{}
		requestBody.Name = cloudflare.String(fullDomain)
		requestBody.Type = cloudflare.F(dns.ARecordType(recordType))
		requestBody.Content = cloudflare.String(originValue)
		requestBody.Comment = cloudflare.String(p.CustomConfig.Comment)
		request := dns.RecordEditParams{
			ZoneID: cloudflare.String(p.cfg.ZoneID),
			Body:   requestBody,
		}
		// 请求
		response, err := p.Provider.Client.DNS.Records.Edit(ctx, p.InternalData.RecordID, request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Cloudfare DNS 自定义配置
type CloudfareDNSCustomConfig struct {
	Comment string `json:"comment"` // 记录集的描述信息
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *CloudfareDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
