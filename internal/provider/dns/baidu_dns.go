package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/baidubce/bce-sdk-go/services/dns"
	"github.com/baidubce/bce-sdk-go/util"
)

// BaiduDNSConfig 百度云 DNS 配置
type BaiduDNSConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	SecretAccessKey string `json:"secret_access_key"`  // 密钥Key
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证百度云 DNS 配置
func (cfg *BaiduDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
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

// NewDefaultBaiduDNSConfig 创建默认的百度云 DNS 配置
func NewDefaultBaiduDNSConfig() *BaiduDNSConfig {
	return &BaiduDNSConfig{
		Endpoint: "dns.baidubce.com",
	}
}

// ParseCustomConfig 解析 百度云 DNS 自定义配置
func (p *BaiduDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// BaiduDNSProvider 百度云 DNS 提供程序
type BaiduDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg  *BaiduDNSConfig
	opts *dnsOptions

	CustomConfig *BaiduDNSCustomConfig
	Provider     struct {
		Client *dns.Client
	}
}

// NewBaiduDNSProvider 创建百度云 DNS 提供程序
func NewBaiduDNSProvider(ctx context.Context, cfg *BaiduDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewBaiduDNSProviderRaw(ctx, cfg, opts...)
}

// NewBaiduDNSProviderRaw 创建百度云 DNS 提供程序
func NewBaiduDNSProviderRaw(ctx context.Context, cfg *BaiduDNSConfig, opts ...DnsOption) (*BaiduDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &BaiduDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		CustomConfig: &BaiduDNSCustomConfig{},
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
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Client.HTTPClient = &http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *BaiduDNSProvider) Name() string {
	return "百度云 DNS"
}

// Check 检查子域名是否存在
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#查询解析记录列表
func (p *BaiduDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (recordID string, isContinue bool, err error) {

	// 请求参数
	request := &dns.ListRecordRequest{}
	request.Rr = subdomain // 主机记录
	request.MaxKeys = 1000 // 每页包含的最大数量，最大数量通常不超过1000，缺省值为1000
	// 请求
	response, err := p.Provider.Client.ListRecord(domain, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Records == nil || len(response.Records) == 0 {
		return "", true, nil
	}
	for _, item := range response.Records {

		if utils.EqualDomainIgnoreCase(item.Rr, subdomain) {

			// 查找记录类型
			itemRecordType := item.Type
			// 查找记录值
			itemOriginValue := item.Value
			// 查找解析线路
			itemLine := item.Line
			// 查找解析记录描述
			itemDescription := item.Description

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			recordID = item.Id

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				return recordID, true, nil
			}

			// 判断解析记录对应的线路
			if !utils.EqualDomainIgnoreCase(itemLine, p.CustomConfig.Line) {
				return recordID, true, nil
			}

			// 判断解析记录描述
			if !utils.EqualDomainIgnoreCase(itemDescription, p.CustomConfig.Description) {
				return recordID, true, nil
			}

			return recordID, false, nil
		}
	}

	// 未找到记录
	return "", true, nil
}

// Update 更新 DNS 记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#添加解析记录
// https://cloud.baidu.com/doc/DNS/s/El4s7lssr#修改解析记录
func (p *BaiduDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValue string, recordType string) (err error) {

	if recordID == "" {
		// 请求参数
		request := &dns.CreateRecordRequest{}
		request.Rr = subdomain                                              // 主机记录
		request.Type = recordType                                           // 解析记录类型
		request.Value = originValue                                         // 记录值
		request.Line = utils.StringToPtr(p.CustomConfig.Line)               // 解析线路
		request.Description = utils.StringToPtr(p.CustomConfig.Description) // 解析记录描述
		// 请求
		err = p.Provider.Client.CreateRecord(domain, request, util.NewUUID())
		printJSON("创建域名记录", request, nil)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := &dns.UpdateRecordRequest{}
		request.Rr = subdomain                                              // 主机记录
		request.Type = recordType                                           // 解析记录类型
		request.Value = originValue                                         // 记录值
		request.Description = utils.StringToPtr(p.CustomConfig.Description) // 解析记录描述
		// 请求
		err = p.Provider.Client.UpdateRecord(domain, recordID, request, util.NewUUID())
		printJSON("更新域名记录", request, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// 百度云 DNS 自定义配置
type BaiduDNSCustomConfig struct {
	Line        string `json:"line"`                  // 解析线路
	Description string `json:"description,omitempty"` // 解析记录描述
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *BaiduDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}
