package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"vibe_ip/internal/helper/sys_env"
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
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// TencentCDNProvider 腾讯云 CDN 提供程序
type TencentCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *TencentCDNConfig
	opts   *cdnOptions
	logger *slog.Logger

	CustomConfig *TencentCDNCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *cdn.Client
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData cdn.BriefDomain
}

// NewTencentCDNProvider 创建腾讯云 CDN 提供程序
func NewTencentCDNProvider(ctx context.Context, logger *slog.Logger, cfg *TencentCDNConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewTencentCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewTencentCDNProviderRaw 创建腾讯云 CDN 提供程序
func NewTencentCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *TencentCDNConfig, opts ...CdnOption) (*TencentCDNProvider, error) {

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
		logger:       logger,
		CustomConfig: &TencentCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
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
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint // 指定域名
	p.Provider.Cfg.NetworkFailureMaxRetries = 3          // 定义最大重试次数

	// 开启日志
	p.Provider.Cfg.Debug = sys_env.GvaEnv.DEBUG

	// 实例化
	p.Provider.Client, err = cdn.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
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
func (p *TencentCDNProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	requestCdnQuery := cdn.NewDescribeDomainsRequest()
	requestCdnQuery.Offset = common.Int64Ptr(0)
	requestCdnQuery.Limit = common.Int64Ptr(1000)
	requestCdnQuery.Filters = make([]*cdn.DomainFilter, 0)
	requestCdnQuery.Filters = append(requestCdnQuery.Filters, &cdn.DomainFilter{
		Name:  common.StringPtr("domain"),
		Value: common.StringPtrs([]string{fullDomain}),
		Fuzzy: common.BoolPtr(false),
	})
	// 请求
	responseCdnQuery, err := p.Provider.Client.DescribeDomainsWithContext(ctx, requestCdnQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeDomainsWithContext 失败 %w", err)
	}
	// 响应参数
	if responseCdnQuery.Response.Domains == nil || len(responseCdnQuery.Response.Domains) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseCdnQuery.Response.Domains {

		// 查找域名
		itemFullDomain := utils.PtrToString(item.Domain)

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录值列表
			itemOriginValues := make([]string, 0, len(item.Origin.Origins))
			for _, origin := range item.Origin.Origins {
				itemOriginValues = append(itemOriginValues, utils.PtrToString(origin))
			}
			// 查找加速域名业务类型
			itemServiceType := utils.PtrToString(item.ServiceType)
			// 查找区域
			itemArea := utils.PtrToString(item.Area)

			// 域名ID
			p.InternalData.RecordID = utils.PtrToString(item.ResourceId)

			// CDN域名
			p.InternalData.CdnDomain = utils.TrimDotsPrt(item.Cname)

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", newOriginValues, itemOriginValues)
				return p.InternalData, nil
			}

			// 判断加速域名业务类型
			if !utils.EqualDomainIgnoreCase(itemServiceType, p.CustomConfig.ServiceType) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("加速域名业务类型不匹配，期望 %v，实际 %v", p.CustomConfig.ServiceType, itemServiceType)
				return p.InternalData, nil
			}

			// 判断区域
			if !utils.EqualDomainIgnoreCase(itemArea, p.CustomConfig.Area) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("区域不匹配，期望 %v，实际 %v", p.CustomConfig.Area, itemArea)
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

// Update 更新 CDN 记录
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=AddCdnDomain
// https://console.cloud.tencent.com/api/explorer?Product=cdn&Version=2018-06-06&Action=UpdateDomainConfig
func (p *TencentCDNProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 转大写
	recordType = strings.ToUpper(recordType)

	// 主源站类型
	originType := recordType
	if len(newOriginValues) > 1 {
		originType = "ip_ipv6"
	} else {
		if recordType == "AAAA" {
			originType = "ipv6"
		}
		if recordType == "A" {
			originType = "ip"
		}
	}

	if p.InternalData.RecordID == "" {

		// 创建加速域名
		requestCdnCreate := cdn.NewAddCdnDomainRequest()
		requestCdnCreate.Domain = common.StringPtr(fullDomain) // 域名
		requestCdnCreate.Origin = &cdn.Origin{
			Origins:            common.StringPtrs(newOriginValues),                  // 源站地址
			OriginType:         common.StringPtr(originType),                        // 源站类型
			OriginPullProtocol: common.StringPtr(p.CustomConfig.OriginPullProtocol), // 回源协议
		}
		requestCdnCreate.ServiceType = common.StringPtr(p.CustomConfig.ServiceType) // 加速域名业务类型
		requestCdnCreate.Area = common.StringPtr(p.CustomConfig.Area)               // 区域
		_, err = p.Provider.Client.AddCdnDomainWithContext(ctx, requestCdnCreate)
		if err != nil {
			return fmt.Errorf("接口 AddCdnDomainWithContext 失败 %w", err)
		}

	} else if p.InternalData.RecordID != "" {

		// 更新加速域名
		requestCdnUpdate := cdn.NewUpdateDomainConfigRequest()
		requestCdnUpdate.Domain = common.StringPtr(fullDomain)                              // 域名
		requestCdnUpdate.ProjectId = common.Int64Ptr(cast.ToInt64(p.InternalData.RecordID)) // 项目 ID
		requestCdnUpdate.Origin = &cdn.Origin{
			Origins:    common.StringPtrs(newOriginValues), // 源站地址
			OriginType: common.StringPtr(originType),       // 源站类型
		}
		_, err = p.Provider.Client.UpdateDomainConfigWithContext(ctx, requestCdnUpdate)
		if err != nil {
			return fmt.Errorf("接口 UpdateDomainConfigWithContext 失败 %w", err)
		}

	}

	return nil
}

// 腾讯云 CDN 自定义配置
type TencentCDNCustomConfig struct {
	ServiceType        string `json:"service_type"`            // 加速域名业务类型
	Area               string `json:"area"`                    // 区域
	OriginPullProtocol string `json:"origin_pull_protocol"`    // 回源协议
	IPGroupCode        string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *TencentCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
