package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	"go.dtapp.net/library/utils/gojson"
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
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// HuaweiDNSProvider 华为云 DNS 提供程序
type HuaweiDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *HuaweiDNSConfig
	opts   *dnsOptions
	logger *slog.Logger

	CustomConfig *HuaweiDNSCustomConfig
	Provider     struct {
		Auth          *auth.BasicCredentials
		Region        *region.Region
		ClientBuilder *core.HcHttpClientBuilder
		Builder       *core.HcHttpClient
		Client        *dns.DnsClient
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData model.ShowRecordSetByZoneResp
}

// NewHuaweiDNSProvider 创建华为云 DNS 提供程序
func NewHuaweiDNSProvider(ctx context.Context, logger *slog.Logger, cfg *HuaweiDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewHuaweiDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewHuaweiDNSProviderRaw 创建华为云 DNS 提供程序
func NewHuaweiDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *HuaweiDNSConfig, opts ...DnsOption) (*HuaweiDNSProvider, error) {

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
		logger:       logger,
		CustomConfig: &HuaweiDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
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

	// 实例化
	p.Provider.Client = dns.NewDnsClient(p.Provider.Builder)

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *HuaweiDNSProvider) Name() string {
	return "华为云 DNS"
}

// printData 打印数据
func (p *HuaweiDNSProvider) printData(ctx context.Context) {
	p.logger.DebugContext(ctx, "数据",
		slog.String("提供商", p.Name()),
		slog.String("内部", gojson.JsonEncodeNoError(p.InternalData)),
	)
	p.logger.DebugContext(ctx, "数据",
		slog.String("提供商", p.Name()),
		slog.String("记录", gojson.JsonEncodeNoError(p.ItemData)),
	)
}

// Check 检查子域名是否存在
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=ShowRecordSetByZone
func (p *HuaweiDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

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
	if response.Recordsets == nil || len(*response.Recordsets) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range *response.Recordsets {

		// 查找域名
		itemFullDomain := utils.PtrToString(item.Name)

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = item

			// 查找记录类型
			itemRecordType := utils.PtrToString(item.Type)
			// 查找记录值
			itemOriginValue := strings.Join(utils.PtrToStrings(item.Records), ",")
			// 查找记录集的描述信息
			itemDescription := utils.PtrToString(item.Description)

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = utils.PtrToString(item.Id)

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", newOriginValue, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录描述
			if !utils.EqualDomainIgnoreCase(itemDescription, p.CustomConfig.Description) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录描述不匹配，期望 %v，实际 %v", p.CustomConfig.Description, itemDescription)
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
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/debug?api=CreateRecordSetWithLine
// https://console.huaweicloud.com/apiexplorer/#/openapi/DNS/sdk?api=UpdateRecordSets
func (p *HuaweiDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 打印数据
	p.printData(ctx)

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if p.InternalData.RecordID == "" {
		// 请求参数
		request := &model.CreateRecordSetWithLineRequest{}
		request.ZoneId = p.cfg.ZoneID // 域名ID
		request.Body = &model.CreateRecordSetWithLineRequestBody{
			Records:     utils.StringsToPtr([]string{originValue}),     // 解析记录的值
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

	if p.InternalData.RecordID != "" {
		// 请求参数
		request := &model.UpdateRecordSetsRequest{}
		request.ZoneId = p.cfg.ZoneID                 // 域名ID
		request.RecordsetId = p.InternalData.RecordID // 记录集ID
		request.Body = &model.UpdateRecordSetsReq{
			Records:     utils.StringsToPtr([]string{originValue}),     // 解析记录的值
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

// 华为云 DNS 自定义配置
type HuaweiDNSCustomConfig struct {
	Description string `json:"description"`             // 记录集的描述信息
	IPGroupCode string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *HuaweiDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
