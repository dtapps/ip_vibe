package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"vibe_ip/internal/provider"

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
func (p *TencentEOProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, cdnDomain string, originValues []string, err error) {

	// // 完整域名
	// fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// // 请求参数
	// request := &alidns20150109.DescribeDomainRecordsRequest{}
	// request.SearchMode = tea.String("COMBINATION") // 搜索模式
	// request.DomainName = tea.String(domain)        // 域名名称
	// request.RRKeyWord = tea.String(subdomain)      // 主机记录的关键字
	// if recordType != "" {
	// 	request.TypeKeyWord = tea.String(recordType) // 解析类型的关键字
	// }
	// // 请求
	// response, err := p.Provider.Client.DescribeDomainRecordsWithContext(ctx, request, &dara.RuntimeOptions{})
	// printJSON("查询加速域名", request, response)
	// if err != nil {
	// 	return
	// }
	// // 响应参数
	// for _, item := range response.Body.DomainRecords.Record {
	// 	if utils.EqualStringIgnoreCase(tea.StringValue(item.RR), subdomain) && utils.EqualStringIgnoreCase(tea.StringValue(item.DomainName), domain) {
	// 		if recordType != "" && !utils.EqualStringIgnoreCase(tea.StringValue(item.Type), recordType) {
	// 			continue
	// 		}
	// 		recordID = tea.StringValue(item.RecordId)
	// 		value = tea.StringValue(item.Value)
	// 		return recordID, value, nil
	// 	}
	// }

	return "", "", []string{}, provider.ErrRecordNotFound
}

// Update 更新 CDN 记录
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=CreateAccelerationDomain
// https://console.cloud.tencent.com/api/explorer?Product=teo&Version=2022-09-01&Action=ModifyAccelerationDomain
func (p *TencentEOProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValues []string, recordType string) (err error) {

	// // 完整域名
	// fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// // 转大写
	// recordType = strings.ToUpper(recordType)
	// originProtocol = strings.ToUpper(originProtocol)

	// // 默认值
	// if httpPort == 0 {
	// 	httpPort = 80
	// }
	// if httpsPort == 0 {
	// 	httpsPort = 443
	// }

	// if recordID == "" {
	// 	// 请求参数
	// 	request := &alidns20150109.AddDomainRecordRequest{}
	// 	request.DomainName = tea.String(domain)        // 域名名称
	// 	request.RR = tea.String(subdomain)             // 主机记录
	// 	request.Type = tea.String(recordType)          // 解析记录类型
	// 	request.Value = tea.String(value)              // 记录值
	// 	request.Line = tea.String(p.CustomConfig.Line) // 解析线路
	// 	// 请求
	// 	response, err := p.Provider.Client.AddDomainRecordWithContext(ctx, request, &dara.RuntimeOptions{})
	// 	printJSON("创建加速域名", request, response)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// if recordID != "" {
	// 	// 请求参数
	// 	request := &alidns20150109.UpdateDomainRecordRequest{}
	// 	request.RecordId = tea.String(recordID)        // 解析记录的ID
	// 	request.RR = tea.String(subdomain)             // 主机记录
	// 	request.Type = tea.String(recordType)          // 解析记录类型
	// 	request.Value = tea.String(value)              // 记录值
	// 	request.Line = tea.String(p.CustomConfig.Line) // 解析线路
	// 	// 请求
	// 	response, err := p.Provider.Client.UpdateDomainRecordWithContext(ctx, request, &dara.RuntimeOptions{})
	// 	printJSON("更新加速域名", request, response)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// 腾讯云 EO 自定义配置
type TencentEOCustomConfig struct {
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
