package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"
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
		Region: "cn-north-1",
	}
}

// ParseCustomConfig 解析 华为云 CDN 自定义配置
func (p *HuaweiCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// HuaweiCDNProvider 华为云 CDN 提供程序
type HuaweiCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *HuaweiCDNConfig
	opts   *cdnOptions
	logger *slog.Logger

	CustomConfig *HuaweiCDNCustomConfig
	Provider     struct {
		Auth          *auth.GlobalCredentials
		Region        *region.Region
		ClientBuilder *core.HcHttpClientBuilder
		Builder       *core.HcHttpClient
		Client        *cdn.CdnClient
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData model.Domains
}

// NewHuaweiCDNProvider 创建华为云 CDN 提供程序
func NewHuaweiCDNProvider(ctx context.Context, logger *slog.Logger, cfg *HuaweiCDNConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewHuaweiCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewHuaweiCDNProviderRaw 创建华为云 CDN 提供程序
func NewHuaweiCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *HuaweiCDNConfig, opts ...CdnOption) (*HuaweiCDNProvider, error) {

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
		logger:       logger,
		CustomConfig: &HuaweiCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth, err = global.NewCredentialsBuilder().
		WithAk(p.cfg.AccessKeyID).
		WithSk(p.cfg.SecretAccessKey).
		SafeBuild()
	if err != nil {
		return nil, fmt.Errorf("创建凭证失败 %w", err)
	}

	// 区域
	p.Provider.Region, err = regionCdn.SafeValueOf(p.cfg.Region)
	if err != nil {
		return nil, fmt.Errorf("创建区域失败 %w", err)
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
		return nil, fmt.Errorf("创建客户端失败 %w", err)
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
func (p *HuaweiCDNProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	requestCdnQuery := &model.ListDomainsRequest{}
	requestCdnQuery.DomainName = utils.StringToPtr(fullDomain) // 加速域名
	requestCdnQuery.PageSize = utils.Int32ToPtr(10000)         // 每页加速域名的数量
	requestCdnQuery.PageNumber = utils.Int32ToPtr(1)           // 查询的页码
	requestCdnQuery.ExactMatch = utils.BoolToPtr(true)         // 精准匹配
	// 请求
	responseCdnQuery, err := p.Provider.Client.ListDomains(requestCdnQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 ListDomains 失败 %w", err)
	}
	// 响应参数
	if responseCdnQuery.Domains == nil || len(*responseCdnQuery.Domains) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range *responseCdnQuery.Domains {

		// 查找域名
		itemFullDomain := utils.PtrToString(item.DomainName)

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = item

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
			sources := *item.Sources
			for _, source := range sources {
				itemOriginValues = append(itemOriginValues, source.IpOrDomain)
			}
			// 查找域名业务类型
			itemBusinessType := utils.PtrToString(item.BusinessType)
			// 查找服务区域
			itemServiceArea := item.ServiceArea.Value()

			// 域名ID
			p.InternalData.RecordID = utils.PtrToString(item.Id)

			// CDN域名
			p.InternalData.CdnDomain = utils.TrimDotsPrt(item.Cname)

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", newOriginValues, itemOriginValues)
				return p.InternalData, nil
			}

			// 判断域名业务类型
			if !utils.EqualDomainIgnoreCase(itemBusinessType, p.CustomConfig.BusinessType) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("域名业务类型不匹配，期望 %v，实际 %v", p.CustomConfig.BusinessType, itemBusinessType)
				return p.InternalData, nil
			}

			// 判断服务区域
			if !utils.EqualDomainIgnoreCase(itemServiceArea, p.CustomConfig.ServiceArea) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("服务区域不匹配，期望 %v，实际 %v", p.CustomConfig.ServiceArea, itemServiceArea)
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
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/debug?version=v2&api=CreateDomain
// https://console.huaweicloud.com/apiexplorer/#/openapi/CDN/sdk?version=v2&api=UpdateDomainFullConfig
func (p *HuaweiCDNProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	var business_type = model.DomainBodyBusinessType{}
	switch p.CustomConfig.BusinessType {
	case "web":
		business_type = model.GetDomainBodyBusinessTypeEnum().WEB
	case "download":
		business_type = model.GetDomainBodyBusinessTypeEnum().DOWNLOAD
	case "video":
		business_type = model.GetDomainBodyBusinessTypeEnum().VIDEO
	case "wholeSite":
		business_type = model.GetDomainBodyBusinessTypeEnum().WHOLE_SITE
	}

	var service_area = model.DomainBodyServiceArea{}
	switch p.CustomConfig.ServiceArea {
	case "mainland_china":
		service_area = model.GetDomainBodyServiceAreaEnum().MAINLAND_CHINA
	case "outside_mainland_china":
		service_area = model.GetDomainBodyServiceAreaEnum().OUTSIDE_MAINLAND_CHINA
	case "global":
		service_area = model.GetDomainBodyServiceAreaEnum().GLOBAL
	}

	if p.InternalData.RecordID == "" {

		sources := make([]model.SourcesRequestBody, 0)
		for _, origin := range newOriginValues {
			sources = append(sources, model.SourcesRequestBody{
				IpOrDomain:    origin,
				OriginType:    model.GetSourcesRequestBodyOriginTypeEnum().IPADDR,
				ActiveStandby: int32(1),
			})
		}

		// 创建加速域名
		requestCdnCreate := &model.CreateDomainRequest{}
		requestCdnCreate.Body = &model.CreateDomainRequestBody{
			Domain: &model.DomainBody{
				DomainName:   fullDomain,
				BusinessType: business_type,
				Sources:      sources,
				ServiceArea:  &service_area,
			},
		}
		responseCdnCreate, err := p.Provider.Client.CreateDomain(requestCdnCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateDomain 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = utils.PtrToString(responseCdnCreate.Domain.Id)

		// CDN域名
		p.InternalData.CdnDomain = utils.TrimDotsPrt(responseCdnCreate.Domain.Cname)

	} else if p.InternalData.RecordID != "" {

		sources := make([]model.SourcesConfig, 0)
		for _, origin := range newOriginValues {
			sources = append(sources, model.SourcesConfig{
				OriginType: model.GetSourcesRequestBodyOriginTypeEnum().IPADDR.Value(),
				OriginAddr: origin,
			})
		}

		// 更新加速域名
		requestCdnUpdate := &model.UpdateDomainFullConfigRequest{}
		requestCdnUpdate.DomainName = fullDomain // 加速域名
		requestCdnUpdate.Body = &model.ModifyDomainConfigRequestBody{
			Configs: &model.Configs{
				BusinessType: utils.StringToPtr(p.CustomConfig.BusinessType),
				ServiceArea:  utils.StringToPtr(p.CustomConfig.ServiceArea),
				Remark:       utils.StringToPtr(p.CustomConfig.Remark),
				Sources:      &sources,
			},
		}
		_, err = p.Provider.Client.UpdateDomainFullConfig(requestCdnUpdate)
		if err != nil {
			return fmt.Errorf("接口 UpdateDomainFullConfig 失败 %w", err)
		}

	}

	return nil
}

// 华为云 CDN 自定义配置
type HuaweiCDNCustomConfig struct {
	BusinessType string `json:"business_type"`           // 域名业务类型
	ServiceArea  string `json:"service_area"`            // 服务区域
	HttpPort     int    `json:"http_port,omitempty"`     // HTTP 端口
	HttpsPort    int    `json:"https_port,omitempty"`    // HTTPS 端口
	Remark       string `json:"remark,omitempty"`        // 备注
	IPGroupCode  string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *HuaweiCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
