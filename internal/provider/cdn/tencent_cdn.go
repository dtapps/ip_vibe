package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	cdn "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cdn/v20180606"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// TencentCDNConfig 腾讯云 CDN 配置
type TencentCDNConfig struct {
	Endpoint   string `json:"endpoint,omitempty"` // 接口地址
	SecretID   string `json:"secret_id"`          // 密钥ID
	SecretKey  string `json:"secret_key"`         // 密钥Key
	DomainName string `json:"domain_name"`        // 主域名
}

// Validate 验证腾讯云 CDN 配置
func (cfg *TencentCDNConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultTencentCDNConfig 创建默认的腾讯云 CDN 配置
func NewDefaultTencentCDNConfig() *TencentCDNConfig {
	return &TencentCDNConfig{
		Endpoint: "cdn.tencentcloudapi.com",
	}
}

// ParseCustomConfig 解析 腾讯云 CDN 自定义配置
func (p *TencentCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// TencentCDNProvider 腾讯云 CDN 提供程序
type TencentCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg  *TencentCDNConfig
	opts *cdnOptions

	CustomConfig *TencentCDNCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *cdn.Client
	}
}

// NewTencentCDNProvider 创建腾讯云 CDN 提供程序
func NewTencentCDNProvider(ctx context.Context, cfg *TencentCDNConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewTencentCDNProviderRaw(ctx, cfg, opts...)
}

// NewTencentCDNProviderRaw 创建腾讯云 CDN 提供程序
func NewTencentCDNProviderRaw(ctx context.Context, cfg *TencentCDNConfig, opts ...CdnOption) (*TencentCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentCDNProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
		CustomConfig: &TencentCDNCustomConfig{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth = common.NewCredential(
		p.cfg.SecretID,
		p.cfg.SecretKey,
	)

	// 选项
	p.Provider.Cfg = profile.NewClientProfile()
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint

	// 实例化
	p.Provider.Client, err = cdn.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Client.WithHttpTransport(p.opts.httpTransport.Instance(nil))
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *TencentCDNProvider) Name() string {
	return "腾讯云 CDN"
}

// Check 检查子域名是否存在
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=DescribeDomains
func (p *TencentCDNProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := cdn.NewDescribeDomainsRequest()
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(1000)
	request.Filters = make([]*cdn.DomainFilter, 0)
	request.Filters = append(request.Filters, &cdn.DomainFilter{
		Name:  common.StringPtr("domain"),
		Value: common.StringPtrs([]string{fullDomain}),
		Fuzzy: common.BoolPtr(false),
	})
	// 请求
	response, err := p.Provider.Client.DescribeDomainsWithContext(ctx, request)
	printJSON("查询加速域名", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Response.Domains {
		if strings.EqualFold(utils.PtrToString(item.Domain), fullDomain) {
			recordID = utils.PtrToString(item.ResourceId)
			value = utils.PtrToString(item.Origin.Origins[0])
			return recordID, value, nil
		}
	}

	return "", "", provider.ErrRecordNotFound
}

// Update 更新 CDN 记录
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=AddCdnDomain
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=UpdateDomainConfig
func (p *TencentCDNProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if recordID == "" {
		// 请求参数
		request := cdn.NewAddCdnDomainRequest()
		request.Domain = common.StringPtr(fullDomain) // 域名
		request.ServiceType = common.StringPtr("web") // 加速域名业务类型
		request.Origin = &cdn.Origin{
			Origins:    common.StringPtrs([]string{value}), // 源站地址
			OriginType: common.StringPtr(recordType),       // 源站类型
		}
		// 请求
		response, err := p.Provider.Client.AddCdnDomainWithContext(ctx, request)
		printJSON("创建加速域名", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := cdn.NewUpdateDomainConfigRequest()
		request.Domain = common.StringPtr(fullDomain)               // 域名
		request.ProjectId = common.Int64Ptr(cast.ToInt64(recordID)) // 项目 ID
		request.Origin = &cdn.Origin{
			Origins:    common.StringPtrs([]string{value}), // 源站地址
			OriginType: common.StringPtr(recordType),       // 源站类型
		}
		// 请求
		response, err := p.Provider.Client.UpdateDomainConfigWithContext(ctx, request)
		printJSON("更新加速域名", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 腾讯云 CDN 自定义配置
type TencentCDNCustomConfig struct {
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *TencentCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
