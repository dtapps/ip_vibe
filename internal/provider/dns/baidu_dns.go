package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"uuid"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/dns"
)

// BaiduDNSConfig 百度云 DNS 配置
type BaiduDNSConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	SecretAccessKey string `json:"secret_access_key"`  // 密钥Key
}

// DefaultValidate 验证配置
func (cfg *BaiduDNSConfig) DefaultValidate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("endpoint 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("SecretAccessKey 不能为空")
	}
	return nil
}

// NewDefaultBaiduDNSConfig 创建默认的百度云 DNS 配置
func NewDefaultBaiduDNSConfig() *BaiduDNSConfig {
	return &BaiduDNSConfig{
		Endpoint: "https://dns.baidubce.com",
	}
}

// ParseCustomConfig 解析 百度云 DNS 自定义配置
func (p *BaiduDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// BaiduDNSProvider 百度云 DNS 提供程序
type BaiduDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *BaiduDNSConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *BaiduDNSCustomConfig
	Provider     struct {
		Client *dns.Client
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData dns.Record
}

// NewBaiduDNSProvider 创建百度云 DNS 提供程序
func NewBaiduDNSProvider(ctx context.Context, logger *slog.Logger, cfg *BaiduDNSConfig, opts ...Option) (provider.DNSProvider, error) {
	return NewBaiduDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewBaiduDNSProviderRaw 创建百度云 DNS 提供程序
func NewBaiduDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *BaiduDNSConfig, opts ...Option) (*BaiduDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &BaiduDNSProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &BaiduDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化
	p.Provider.Client, err = dns.NewClient(
		p.cfg.AccessKeyID,
		p.cfg.SecretAccessKey,
		p.cfg.Endpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = dnsHttpTransport
	}
	if p.opts.httpTransport != nil {
		p.Provider.Client.HTTPClient = &http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}
	}

	// 定义最大重试次数
	p.Provider.Client.Config.Retry = bce.NewBackOffRetryPolicy(3, 20000, 300)

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *BaiduDNSProvider) Name() string {
	return "百度云 DNS"
}

// Check 检查子域名是否存在
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#查询解析记录列表
func (p *BaiduDNSProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (data *provider.DNSProviderInternalData, err error) {

	// 查询域名记录
	requestDnsQuery := &dns.ListRecordRequest{}
	requestDnsQuery.Rr = sub_domain // 主机记录
	requestDnsQuery.MaxKeys = 1000  // 每页包含的最大数量，最大数量通常不超过1000，缺省值为1000
	responseDnsQuery, err := p.Provider.Client.ListRecord(domain_name, requestDnsQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 ListRecord 失败 %w", err)
	}
	// 响应参数
	if len(responseDnsQuery.Records) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseDnsQuery.Records {

		if utils.EqualDomainIgnoreCase(item.Rr, sub_domain) {

			// 记录数据
			p.ItemData = item

			// 查找记录类型
			itemRecordType := item.Type
			// 查找记录值
			itemOriginValue := item.Value
			// 查找解析线路
			itemLine := item.Line
			// 查找解析记录描述
			itemDescription := item.Description

			// 检查记录类型
			if record_type != "" && !utils.EqualDomainIgnoreCase(itemRecordType, record_type) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = item.Id

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, origin_value) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", origin_value, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录对应的线路
			if !utils.EqualDomainIgnoreCase(itemLine, p.CustomConfig.LineValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录对应的线路不匹配，期望 %v，实际 %v", p.CustomConfig.LineValue, itemLine)
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
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#添加解析记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#修改解析记录
func (p *BaiduDNSProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (err error) {

	if p.InternalData.RecordID == "" {

		// 创建域名记录
		requestDnsCreate := &dns.CreateRecordRequest{}
		requestDnsCreate.Rr = sub_domain                                             // 主机记录
		requestDnsCreate.Type = record_type                                          // 解析记录类型
		requestDnsCreate.Value = origin_value                                        // 记录值
		requestDnsCreate.Line = utils.StringToPtr(p.CustomConfig.LineValue)          // 解析线路
		requestDnsCreate.Description = utils.StringToPtr(p.CustomConfig.Description) // 解析记录描述
		err = p.Provider.Client.CreateRecord(domain_name, requestDnsCreate, uuid.NewV7().String())
		if err != nil {
			return fmt.Errorf("接口 CreateRecord 失败 %w", err)
		}

	} else if p.InternalData.RecordID != "" {

		// 更新域名记录
		requestDnsUpdate := &dns.UpdateRecordRequest{}
		requestDnsUpdate.Rr = sub_domain                                             // 主机记录
		requestDnsUpdate.Type = record_type                                          // 解析记录类型
		requestDnsUpdate.Value = origin_value                                        // 记录值
		requestDnsUpdate.Description = utils.StringToPtr(p.CustomConfig.Description) // 解析记录描述
		err = p.Provider.Client.UpdateRecord(domain_name, p.InternalData.RecordID, requestDnsUpdate, uuid.NewV7().String())
		if err != nil {
			return fmt.Errorf("接口 UpdateRecord 失败 %w", err)
		}

	}

	return nil
}

// 百度云 DNS 自定义配置
type BaiduDNSCustomConfig struct {
	ZoneValue     string `json:"zone_value"`                 // 域名编号
	LineValue     string `json:"line_value"`                 // 线路
	IPGroupIDCode string `json:"ip_group_id_code,omitempty"` // 设备IP组代码
	Description   string `json:"description,omitempty"`      // 备注
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *BaiduDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// 查询域名列表
// https://cloud.baidu.com/doc/DNS/s/kl4s7g11z#查询域名列表
func (p *BaiduDNSProvider) ListZone(ctx context.Context) ([]dns.Zone, error) {
	// 查询线路列表
	request := &dns.ListZoneRequest{}
	response, err := p.Provider.Client.ListZone(request)
	if err != nil {
		return nil, fmt.Errorf("接口 ListZone 失败 %w", err)
	}
	return response.Zones, nil
}
