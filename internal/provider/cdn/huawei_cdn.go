package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/config"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/core/region"
	cdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/model"
	regionCdn "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cdn/v2/region"
)

// HuaweiCDNConfig 华为云 CDN 配置
type HuaweiCDNConfig struct {
	Region          string `json:"region,omitempty"`  // 区域
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥密钥
	DomainName      string `json:"domain_name"`       // 主域名
}

// Validate 验证华为云 CDN 配置
func (cfg *HuaweiCDNConfig) Validate() error {
	if cfg.Region == "" {
		return fmt.Errorf("Region 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("SecretAccessKey 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultHuaweiCDNConfig 创建默认的华为云 CDN 配置
func NewDefaultHuaweiCDNConfig() *HuaweiCDNConfig {
	return &HuaweiCDNConfig{
		Region: "cn-south-1",
	}
}

// ParseCustomConfig 解析 华为云 CDN 自定义配置
func (p *HuaweiCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// HuaweiCDNProvider 华为云 CDN 提供程序
type HuaweiCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg  *HuaweiCDNConfig
	opts *cdnOptions

	CustomConfig *HuaweiCDNCustomConfig
	Provider     struct {
		Auth          *auth.BasicCredentials
		Region        *region.Region
		ClientBuilder *core.HcHttpClientBuilder
		Builder       *core.HcHttpClient
		Client        *cdn.CdnClient
	}
}

// NewHuaweiCDNProvider 创建华为云 CDN 提供程序
func NewHuaweiCDNProvider(ctx context.Context, cfg *HuaweiCDNConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewHuaweiCDNProviderRaw(ctx, cfg, opts...)
}

// NewHuaweiCDNProviderRaw 创建华为云 CDN 提供程序
func NewHuaweiCDNProviderRaw(ctx context.Context, cfg *HuaweiCDNConfig, opts ...CdnOption) (*HuaweiCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &HuaweiCDNProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
		CustomConfig: &HuaweiCDNCustomConfig{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth, err = basic.NewCredentialsBuilder().
		WithAk(p.cfg.AccessKeyID).
		WithSk(p.cfg.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	// 区域
	p.Provider.Region, err = regionCdn.SafeValueOf(p.cfg.Region)
	if err != nil {
		return nil, err
	}

	p.Provider.ClientBuilder = cdn.CdnClientBuilder().
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

	// 实例化
	p.Provider.Client = cdn.NewCdnClient(p.Provider.Builder)

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *HuaweiCDNProvider) Name() string {
	return "华为云 CDN"
}

// Check 检查子域名是否存在
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/sdk?version=v2&api=ListDomains
func (p *HuaweiCDNProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, cdnDomain string, originValues []string, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := &model.ListDomainsRequest{}
	request.DomainName = utils.StringToPtr(fullDomain) // 加速域名
	request.PageSize = utils.Int32ToPtr(10000)         // 每页加速域名的数量
	request.PageNumber = utils.Int32ToPtr(1)           // 查询的页码
	request.ExactMatch = utils.BoolToPtr(true)         // 精准匹配
	// 请求
	response, err := p.Provider.Client.ListDomains(request)
	printJSON("查询加速域名", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range *response.Domains {
		if utils.EqualStringIgnoreCase(utils.TrimDots(utils.PtrToString(item.DomainName)), utils.TrimDots(fullDomain)) {
			// 域名ID
			recordID = utils.PtrToString(item.Id)
			// CDN域名
			cdnDomain = utils.TrimDots(utils.PtrToString(item.DomainName))
			// 记录值列表
			sources := *item.Sources
			for _, source := range sources {
				originValues = append(originValues, source.IpOrDomain)
			}
			return recordID, cdnDomain, originValues, nil
		}
	}

	return "", "", []string{}, provider.ErrRecordNotFound
}

// Update 更新 CDN 记录
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/debug?version=v2&api=CreateDomain
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/sdk?version=v2&api=UpdateDomainFullConfig
func (p *HuaweiCDNProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if recordID == "" {
		// 请求参数
		request := &model.CreateDomainRequest{}
		request.Body = &model.CreateDomainRequestBody{
			Domain: &model.DomainBody{
				DomainName:   fullDomain,
				BusinessType: model.GetDomainBodyBusinessTypeEnum().WEB,
				Sources: []model.SourcesRequestBody{
					{
						IpOrDomain:    originValues[0],
						OriginType:    model.GetSourcesRequestBodyOriginTypeEnum().IPADDR,
						ActiveStandby: int32(1),
					},
				},
			},
		}
		// 请求
		response, err := p.Provider.Client.CreateDomain(request)
		printJSON("创建加速域名", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := &model.UpdateDomainFullConfigRequest{}
		request.DomainName = fullDomain // 加速域名
		request.Body = &model.ModifyDomainConfigRequestBody{
			Configs: &model.Configs{
				Sources: &[]model.SourcesConfig{
					{
						OriginType: model.GetSourcesRequestBodyOriginTypeEnum().IPADDR.Value(),
						OriginAddr: originValues[0],
					},
				},
			},
		}
		// 请求
		response, err := p.Provider.Client.UpdateDomainFullConfig(request)
		printJSON("更新加速域名", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 华为云 CDN 自定义配置
type HuaweiCDNCustomConfig struct {
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *HuaweiCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
