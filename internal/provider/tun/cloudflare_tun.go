package tun

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"vibe_ip/internal/provider"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/zero_trust"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

// CloudfareTUNConfig Cloudfare TUN 配置
type CloudfareTUNConfig struct {
	ApiToken  string `json:"api_token"`  // API令牌
	AccountID string `json:"account_id"` // 账号ID
}

// Validate 验证Cloudfare TUN 配置
func (cfg *CloudfareTUNConfig) Validate() error {
	if cfg.ApiToken == "" {
		return fmt.Errorf("ApiToken 不能为空")
	}
	if cfg.AccountID == "" {
		return fmt.Errorf("AccountID 不能为空")
	}
	return nil
}

// NewDefaultCloudfareTUNConfig 创建默认的Cloudfare TUN 配置
func NewDefaultCloudfareTUNConfig() *CloudfareTUNConfig {
	return &CloudfareTUNConfig{}
}

// ParseCustomConfig 解析 Cloudfare TUN 自定义配置
func (p *CloudfareTUNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// CloudfareTUNProvider Cloudfare TUN 提供程序
type CloudfareTUNProvider struct {
	provider.UnimplementedTUNProvider // 嵌入未实现的

	cfg    *CloudfareTUNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *CloudfareTUNCustomConfig
	Provider     struct {
		Opts   []option.RequestOption
		Client *cloudflare.Client
	}

	// 内部数据
	InternalData *provider.TUNProviderInternalData
}

// NewCloudfareTUNProvider 创建Cloudfare TUN 提供程序
func NewCloudfareTUNProvider(ctx context.Context, logger *slog.Logger, cfg *CloudfareTUNConfig, opts ...Option) (provider.TUNProvider, error) {
	return NewCloudfareTUNProviderRaw(ctx, logger, cfg, opts...)
}

// NewCloudfareTUNProviderRaw 创建Cloudfare TUN 提供程序
func NewCloudfareTUNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *CloudfareTUNConfig, opts ...Option) (*CloudfareTUNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &CloudfareTUNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &CloudfareTUNCustomConfig{},
		InternalData: &provider.TUNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化client选项
	p.Provider.Opts = append(p.Provider.Opts, option.WithAPIToken(p.cfg.ApiToken))

	// 定义最大重试次数
	p.Provider.Opts = append(p.Provider.Opts, option.WithMaxRetries(3))

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
func (p *CloudfareTUNProvider) Name() string {
	return "Cloudflare TUN"
}

// Check 检查子域名是否存在
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/list/
func (p *CloudfareTUNProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.TUNProviderInternalData, err error) {

	// // 完整域名
	// fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// // 查询域名记录
	// requestDnsQuery := dns.RecordListParams{}
	// requestDnsQuery.ZoneID = cloudflare.String(p.cfg.ZoneID)
	// requestDnsQuery.Name = cloudflare.F(dns.RecordListParamsName{
	// 	Exact: cloudflare.String(fullDomain),
	// })
	// if recordType != "" {
	// 	requestDnsQuery.Type = cloudflare.F(dns.RecordListParamsType(recordType))
	// }
	// responseDnsQuery, err := p.Provider.Client.DNS.Records.List(ctx, requestDnsQuery)
	// if err != nil {
	// 	return nil, fmt.Errorf("接口 DNS.Records.List 失败 %w", err)
	// }
	// // 响应参数
	// if responseDnsQuery.Result == nil || len(responseDnsQuery.Result) == 0 {
	// 	p.InternalData.IsContinue = true
	// 	return p.InternalData, nil
	// }
	// for _, item := range responseDnsQuery.Result {

	// 	// 查找域名
	// 	itemFullDomain := item.Name

	// 	if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

	// 		// 记录数据
	// 		p.ItemData = item

	// 		// 查找记录类型
	// 		itemRecordType := cast.ToString(item.Type)
	// 		// 查找记录值
	// 		itemOriginValue := item.Content
	// 		// 查找记录集的描述信息
	// 		itemComment := item.Comment

	// 		// 检查记录类型
	// 		if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
	// 			continue
	// 		}

	// 		// 域名ID
	// 		p.InternalData.RecordID = item.ID

	// 		// 判断记录值
	// 		if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
	// 			p.InternalData.IsContinue = true
	// 			p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", newOriginValue, itemOriginValue)
	// 			return p.InternalData, nil
	// 		}

	// 		// 判断解析记录描述
	// 		if !utils.EqualDomainIgnoreCase(itemComment, p.CustomConfig.Comment) {
	// 			p.InternalData.IsContinue = true
	// 			p.InternalData.Tips = fmt.Sprintf("解析记录描述不匹配，期望 %v，实际 %v", p.CustomConfig.Comment, itemComment)
	// 			return p.InternalData, nil
	// 		}

	// 		p.InternalData.IsContinue = false
	// 		return p.InternalData, nil
	// 	}
	// }

	// // 未找到记录
	// p.InternalData.IsContinue = true
	return p.InternalData, nil
}

// Update 更新 DNS 记录
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/create/
// https://developers.cloudflare.com/api/go/resources/dns/subresources/records/methods/edit/
func (p *CloudfareTUNProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// // 完整域名
	// fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// if p.InternalData.RecordID == "" {

	// 	// 创建域名记录
	// 	requestDnsCreateBody := dns.ARecordParam{}
	// 	requestDnsCreateBody.Name = cloudflare.String(fullDomain)
	// 	requestDnsCreateBody.Type = cloudflare.F(dns.ARecordType(recordType))
	// 	requestDnsCreateBody.Content = cloudflare.String(originValue)
	// 	requestDnsCreateBody.Comment = cloudflare.String(p.CustomConfig.Comment)
	// 	requestDnsCreate := dns.RecordNewParams{
	// 		ZoneID: cloudflare.String(p.cfg.ZoneID),
	// 		Body:   requestDnsCreateBody,
	// 	}
	// 	responseDnsCreate, err := p.Provider.Client.DNS.Records.New(ctx, requestDnsCreate)
	// 	if err != nil {
	// 		return fmt.Errorf("接口 DNS.Records.New 失败 %w", err)
	// 	}

	// 	// 域名ID
	// 	p.InternalData.RecordID = responseDnsCreate.ID

	// } else if p.InternalData.RecordID != "" {

	// 	// 更新域名记录
	// 	requestDnsUpdateBody := dns.ARecordParam{}
	// 	requestDnsUpdateBody.Name = cloudflare.String(fullDomain)
	// 	requestDnsUpdateBody.Type = cloudflare.F(dns.ARecordType(recordType))
	// 	requestDnsUpdateBody.Content = cloudflare.String(originValue)
	// 	requestDnsUpdateBody.Comment = cloudflare.String(p.CustomConfig.Comment)
	// 	requestDnsUpdate := dns.RecordEditParams{
	// 		ZoneID: cloudflare.String(p.cfg.ZoneID),
	// 		Body:   requestDnsUpdateBody,
	// 	}
	// 	responseDnsUpdate, err := p.Provider.Client.DNS.Records.Edit(ctx, p.InternalData.RecordID, requestDnsUpdate)
	// 	if err != nil {
	// 		return fmt.Errorf("接口 DNS.Records.Edit 失败 %w", err)
	// 	}

	// 	// 域名ID
	// 	p.InternalData.RecordID = responseDnsUpdate.ID

	// }

	return nil
}

// Cloudfare TUN 自定义配置
type CloudfareTUNCustomConfig struct {
	Comment     string `json:"comment"`                 // 记录集的描述信息
	IPGroupCode string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 TUN 自定义配置
func (p *CloudfareTUNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// CloudfareTUNProviderConnector Cloudfare TUN 隧道
type CloudfareTUNProviderConnector struct {
	ID   string `json:"id"`   // ID
	Name string `json:"name"` // 名称
}

// Lists 获取 Cloudfare TUN 隧道列表
// https://developers.cloudflare.com/api/go/resources/zero_trust/subresources/tunnels/methods/list/
func (p *CloudfareTUNProvider) Connectors(ctx context.Context) (lines []*CloudfareTUNProviderConnector, err error) {
	// 查询隧道列表
	requestQueryList := zero_trust.TunnelListParams{}
	requestQueryList.AccountID = cloudflare.F(p.cfg.AccountID)
	requestQueryList.Page = cloudflare.F(float64(1))
	requestQueryList.PerPage = cloudflare.F(float64(1000))
	responseQueryList, err := p.Provider.Client.ZeroTrust.Tunnels.List(ctx, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 ZeroTrust.Tunnels.List 失败 %w", err)
	}
	for _, item := range responseQueryList.Result {
		lines = append(lines, &CloudfareTUNProviderConnector{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return lines, nil
}

// CloudfareTUNProviderZone Cloudfare TUN 域名
type CloudfareTUNProviderZone struct {
	ID   string `json:"id"`   // ID
	Name string `json:"name"` // 名称
}

// Lists 获取 Cloudfare TUN 域名列表
// https://developers.cloudflare.com/api/go/resources/zones/methods/list/
func (p *CloudfareTUNProvider) Zones(ctx context.Context) (lines []*CloudfareTUNProviderZone, err error) {
	// 查询域名列表
	requestQueryList := zones.ZoneListParams{}
	requestQueryList.Page = cloudflare.F(float64(1))
	requestQueryList.PerPage = cloudflare.F(float64(50))
	responseQueryList, err := p.Provider.Client.Zones.List(ctx, requestQueryList)
	if err != nil {
		return nil, fmt.Errorf("接口 Zones.List 失败 %w", err)
	}
	for _, item := range responseQueryList.Result {
		lines = append(lines, &CloudfareTUNProviderZone{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return lines, nil
}
