package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

// TencentEOConfig 腾讯云 EO 配置
type TencentEOConfig struct {
	Endpoint   string `json:"endpoint,omitempty"` // 接口地址
	SecretID   string `json:"secret_id"`          // 密钥ID
	SecretKey  string `json:"secret_key"`         // 密钥Key
	ZoneID     string `json:"zone_id"`            // 区域ID
	DomainName string `json:"domain_name"`        // 主域名
}

// Validate 验证腾讯云 EO 配置
func (cfg *TencentEOConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.ZoneID == "" {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultTencentEOConfig 创建默认的腾讯云 EO 配置
func NewDefaultTencentEOConfig() *TencentEOConfig {
	return &TencentEOConfig{
		Endpoint: "teo.tencentcloudapi.com",
	}
}

// ParseCustomConfig 解析 腾讯云 EO 自定义配置
func (p *TencentEOConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// TencentEOProvider 腾讯云 EO 提供程序
type TencentEOProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg  *TencentEOConfig
	opts *cdnOptions

	CustomConfig *TencentEOCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *teo.Client
	}
}

// NewTencentEOProvider 创建腾讯云 EO 提供程序
func NewTencentEOProvider(ctx context.Context, cfg *TencentEOConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewTencentEOProviderRaw(ctx, cfg, opts...)
}

// NewTencentEOProviderRaw 创建腾讯云 EO 提供程序
func NewTencentEOProviderRaw(ctx context.Context, cfg *TencentEOConfig, opts ...CdnOption) (*TencentEOProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentEOProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
		CustomConfig: &TencentEOCustomConfig{},
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
	p.Provider.Client, err = teo.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
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
func (p *TencentEOProvider) Name() string {
	return "腾讯云 EO"
}

// Check 检查子域名是否存在
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=DescribeAccelerationDomains
func (p *TencentEOProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (recordID string, cdnDomain string, isContinue bool, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := teo.NewDescribeAccelerationDomainsRequest()
	request.ZoneId = common.StringPtr(p.cfg.ZoneID)
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(200)
	request.Filters = make([]*teo.AdvancedFilter, 0)
	request.Filters = append(request.Filters, &teo.AdvancedFilter{
		Name:   common.StringPtr("domain-name"),
		Values: common.StringPtrs([]string{fullDomain}),
	})
	// 请求
	response, err := p.Provider.Client.DescribeAccelerationDomainsWithContext(ctx, request)
	printJSON("查询加速域名", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Response == nil || len(response.Response.AccelerationDomains) == 0 {
		return "", "", true, nil
	}

	// 自定义一个结构体，获取域名ID
	var customResponse struct {
		Response struct {
			AccelerationDomains []struct {
				DomainId string `json:"DomainId,omitempty"`
			} `json:"AccelerationDomains,omitempty"`
		} `json:"Response,omitempty"`
	}
	if err := json.Unmarshal([]byte(response.ToJsonString()), &customResponse); err != nil {
		return "", "", true, err
	}

	for key, item := range response.Response.AccelerationDomains {

		// 查找域名
		itemFullDomain := utils.PtrToString(item.DomainName)

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 查找记录值列表
			itemOriginValues := make([]string, 0)
			itemOriginValues = append(itemOriginValues, utils.PtrToString(item.OriginDetail.Origin))
			// 查找回源协议
			itemOriginProtocol := utils.PtrToString(item.OriginProtocol)
			// 查找HTTP回源端口
			itemHttpPort := utils.PtrToUint64(item.HttpOriginPort)
			// 查找HTTPS回源端口
			itemHttpsPort := utils.PtrToUint64(item.HttpsOriginPort)
			// 查找IPv6状态
			itemIPv6Status := utils.PtrToString(item.IPv6Status)

			// 域名ID
			recordID = customResponse.Response.AccelerationDomains[key].DomainId

			// CDN域名
			cdnDomain = utils.TrimDotsPrt(item.Cname)

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, newOriginValues) {
				return recordID, cdnDomain, true, nil
			}

			// 判断回源协议
			if p.CustomConfig.OriginProtocol != itemOriginProtocol {
				return recordID, cdnDomain, true, nil
			}

			// 判断HTTP回源端口
			if p.CustomConfig.HttpPort != itemHttpPort {
				return recordID, cdnDomain, true, nil
			}

			// 判断HTTPS回源端口
			if p.CustomConfig.HttpsPort != itemHttpsPort {
				return recordID, cdnDomain, true, nil
			}

			// 判断IPv6状态
			if p.CustomConfig.IPv6Status != itemIPv6Status {
				return recordID, cdnDomain, true, nil
			}

			return recordID, cdnDomain, false, nil
		}
	}

	return "", "", true, nil
}

// Update 更新 CDN 记录
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateAccelerationDomain
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyAccelerationDomain
func (p *TencentEOProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 回源协议，转大写
	p.CustomConfig.OriginProtocol = strings.ToUpper(p.CustomConfig.OriginProtocol)

	// 源站信息
	originInfo := new(teo.OriginInfo{})
	if len(originValues) == 1 {
		originInfo.OriginType = common.StringPtr("IP_DOMAIN") // 源站类型
		originInfo.Origin = common.StringPtr(originValues[0]) // 源站地址
	} else {

	}

	// HTTP回源端口，默认80
	if p.CustomConfig.HttpPort == 0 {
		p.CustomConfig.HttpPort = 80
	}

	// HTTPS回源端口，默认443
	if p.CustomConfig.HttpsPort == 0 {
		p.CustomConfig.HttpsPort = 443
	}

	if recordID == "" {
		// 请求参数
		request := teo.NewCreateAccelerationDomainRequest()
		request.ZoneId = common.StringPtr(p.cfg.ZoneID)                              // 加速域名所属站点 ID
		request.DomainName = common.StringPtr(fullDomain)                            // 加速域名
		request.OriginInfo = originInfo                                              // 源站信息
		request.OriginProtocol = common.StringPtr(p.CustomConfig.OriginProtocol)     // 回源协议
		request.HttpOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpPort))   // HTTP回源端口
		request.HttpsOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpsPort)) // HTTPS回源端口
		request.IPv6Status = common.StringPtr(p.CustomConfig.IPv6Status)             // IPv6状态
		// 请求
		response, err := p.Provider.Client.CreateAccelerationDomainWithContext(ctx, request)
		printJSON("创建加速域名", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := teo.NewModifyAccelerationDomainRequest()
		request.ZoneId = common.StringPtr(p.cfg.ZoneID)                              // 加速域名所属站点 ID
		request.DomainName = common.StringPtr(fullDomain)                            // 加速域名
		request.OriginInfo = originInfo                                              // 源站信息
		request.OriginProtocol = common.StringPtr(p.CustomConfig.OriginProtocol)     // 回源协议
		request.HttpOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpPort))   // HTTP回源端口
		request.HttpsOriginPort = common.Uint64Ptr(uint64(p.CustomConfig.HttpsPort)) // HTTPS回源端口
		request.IPv6Status = common.StringPtr(p.CustomConfig.IPv6Status)             // IPv6状态
		// 请求
		response, err := p.Provider.Client.ModifyAccelerationDomainWithContext(ctx, request)
		printJSON("更新加速域名", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 腾讯云 EO 自定义配置
type TencentEOCustomConfig struct {
	OriginProtocol string `json:"origin_protocol"`       // 回源协议
	HttpPort       uint64 `json:"http_port,omitempty"`   // HTTP回源端口
	HttpsPort      uint64 `json:"https_port,omitempty"`  // HTTPS回源端口
	IPv6Status     string `json:"ipv6_status,omitempty"` // IPv6状态
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *TencentEOProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
