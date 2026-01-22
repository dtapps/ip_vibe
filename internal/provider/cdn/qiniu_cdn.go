package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/cdn"
	"go.dtapp.net/library/service/qiniu"
	"resty.dev/v3"
)

// QiniuCDNConfig 七牛云 CDN 配置
type QiniuCDNConfig struct {
	Endpoint   string `json:"endpoint,omitempty"` // 接口地址
	AccessKey  string `json:"access_key"`         // 密钥ID
	SecretKey  string `json:"secret_key"`         // 密钥Key
	DomainName string `json:"domain_name"`        // 主域名
}

// Validate 验证七牛云 CDN 配置
func (cfg *QiniuCDNConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.AccessKey == "" {
		return fmt.Errorf("AccessKey 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultQiniuCDNConfig 创建默认的七牛云 CDN 配置
func NewDefaultQiniuCDNConfig() *QiniuCDNConfig {
	return &QiniuCDNConfig{
		Endpoint: "https://api.qiniu.com",
	}
}

// ParseCustomConfig 解析 七牛云 CDN 自定义配置
func (p *QiniuCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// QiniuCDNProvider 七牛云 CDN 提供程序
type QiniuCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *QiniuCDNConfig
	opts   *cdnOptions
	logger *slog.Logger

	CustomConfig *QiniuCDNCustomConfig
	Provider     struct {
		Auth          *auth.Credentials
		Client        *cdn.CdnManager
		SignTransport *qiniu.SignTransport
		Resty         *resty.Client
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData
}

// NewQiniuCDNProvider 创建七牛云 CDN 提供程序
func NewQiniuCDNProvider(ctx context.Context, logger *slog.Logger, cfg *QiniuCDNConfig, opts ...CdnOption) (provider.CDNProvider, error) {
	return NewQiniuCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewQiniuCDNProviderRaw 创建七牛云 CDN 提供程序
func NewQiniuCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *QiniuCDNConfig, opts ...CdnOption) (*QiniuCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &QiniuCDNProvider{
		cfg:          cfg,
		opts:         &cdnOptions{},
		logger:       logger,
		CustomConfig: &QiniuCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth = auth.New(p.cfg.AccessKey, p.cfg.SecretKey)

	// 实例化
	p.Provider.Client = cdn.NewCdnManager(p.Provider.Auth)

	// 签名信息
	p.Provider.SignTransport = qiniu.NewSignTransport(nil, p.cfg.AccessKey, p.cfg.SecretKey)

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.SignTransport = qiniu.NewSignTransport(p.opts.httpTransport, p.cfg.AccessKey, p.cfg.SecretKey)
	}

	// 实例化Resty客户端
	p.Provider.Resty = resty.New().SetBaseURL(p.cfg.Endpoint)

	// 注入签名信息
	p.Provider.Resty.SetTransport(p.Provider.SignTransport)

	// 定义最大重试次数
	p.Provider.Resty.SetRetryCount(3)

	// 开启日志
	p.Provider.Resty.SetDebug(sys_env.GvaEnv.DEBUG)

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *QiniuCDNProvider) Name() string {
	return "七牛云 CDN"
}

type QiniuCDNProviderCheck struct {
	Domains []struct {
		Name     string `json:"name"`     // 加速域名
		Type     string `json:"type"`     // 域名类型
		Cname    string `json:"cname"`    // 别名
		Protocol string `json:"protocol"` // 协议
		Platform string `json:"platform"` // 平台类型
		GeoCover string `json:"geoCover"` // 地域
	} `json:"domains"`
}

// Check 检查子域名是否存在
// https://developer.qiniu.com/fusion/4246/the-domain-name#11
func (p *QiniuCDNProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 获取域名列表
	var responseCdnList QiniuCDNProviderCheck
	_, err = p.Provider.Resty.R().
		SetContext(ctx).
		SetQueryParam("types", "normal").
		SetQueryParam("marker", "").
		SetQueryParam("limit", "1000").
		SetResult(&responseCdnList).
		Get("/domain")
	if err != nil {
		return nil, fmt.Errorf("接口 get domain 失败 %w", err)
	}
	// 响应参数
	if responseCdnList.Domains == nil || len(responseCdnList.Domains) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseCdnList.Domains {

		// 查找域名
		itemFullDomain := item.Name

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 查找平台类型
			itemPlatform := item.Platform
			// 查找地域
			itemGeoCover := item.GeoCover

			// 域名ID
			// p.InternalData.RecordID = item.ResourceId

			// CDN域名
			p.InternalData.CdnDomain = item.Cname

			// 判断平台类型
			if !utils.EqualStringIgnoreCase(itemPlatform, p.CustomConfig.Platform) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("平台类型不匹配，期望 %v，实际 %v", p.CustomConfig.Platform, itemPlatform)
				return p.InternalData, nil
			}

			// 判断地域
			if !utils.EqualStringIgnoreCase(itemGeoCover, p.CustomConfig.GeoCover) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("地域不匹配，期望 %v，实际 %v", p.CustomConfig.GeoCover, itemGeoCover)
				return p.InternalData, nil
			}

			p.InternalData.IsContinue = false
			return p.InternalData, nil
		}

	}

	// 获取域名信息
	_, err = p.Provider.Resty.R().
		SetContext(ctx).
		SetPathParam("domain", fullDomain).
		Get("/domain/{domain}")
	if err != nil {
		// 未找到记录
		p.InternalData.IsContinue = true
		return p.InternalData, nil
		// return nil, fmt.Errorf("接口 get domain/{domain} 失败 %w", err)
	}

	p.InternalData.IsContinue = false
	return p.InternalData, nil
}

// Update 更新 CDN 记录
// https://developer.qiniu.com/fusion/4246/the-domain-name#1
// https://developer.qiniu.com/fusion/4246/the-domain-name#4
func (p *QiniuCDNProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 回源协议
	if p.CustomConfig.SourceURLScheme == "follow" {
		p.CustomConfig.SourceURLScheme = ""
	}

	// 源站配置
	sources := map[string]any{
		// "sourceType":      "advanced",                     // 源站类型
		// "sourceURLScheme": p.CustomConfig.SourceURLScheme, // 回源协议
		// "advancedSources": []map[string]any{
		// 	{
		// 		"addr":   newOriginValues[0], // 高级回源的回源地址
		// 		"weight": 100,                // 高级回源的回源addr权重
		// 		"backup": false,              // 高级回源的回源addr是否为备源地址
		// 	},
		// },
		"sourceType":      "ip",                           // 源站类型
		"sourceURLScheme": p.CustomConfig.SourceURLScheme, // 回源协议
		"sourceIPs":       newOriginValues,                // 回源ip
		"testURLPath":     "/",                            // 用于测试的URL Path, 检测源站是否可访问
	}

	// 缓存配置
	cache := map[string]any{
		"cacheControls": []map[string]any{
			{
				"time":     0,
				"timeunit": 0,
				"type":     "follow", // 缓存类型
				"rule":     "*",      // 缓存路径规则
			},
		},
		"ignoreParam": false, // 是否忽略参数
	}

	if p.InternalData.CdnDomain == "" {

		// 创建域名
		_, err = p.Provider.Resty.R().
			SetContext(ctx).
			SetBody(map[string]any{
				"type":     "normal",                // 域名类型
				"platform": p.CustomConfig.Platform, // 平台类型
				"geoCover": p.CustomConfig.GeoCover, // 地域
				"protocol": "http",                  // 协议
				"source":   sources,                 // 源站配置
				"cache":    cache,                   // 缓存配置
				"ipTypes":  3,                       // IP协议
			}).
			SetPathParam("domain", fullDomain).
			Post("/domain/{domain}")
		if err != nil {
			return fmt.Errorf("接口 post domain/{domain} 失败 %w", err)
		}

	} else if p.InternalData.CdnDomain != "" {

		// 修改源站
		_, err = p.Provider.Resty.R().
			SetContext(ctx).
			SetBody(map[string]any{
				"source": sources, // 源站配置
			}).
			SetPathParam("domain", fullDomain).
			Put("/domain/{domain}/source")
		if err != nil {
			return fmt.Errorf("接口 put domain/{domain}/source 失败 %w", err)
		}

	}

	return nil
}

// 七牛云 CDN 自定义配置
type QiniuCDNCustomConfig struct {
	Platform        string `json:"platform"`                // 平台类型
	GeoCover        string `json:"geo_cover"`               // 地域
	SourceURLScheme string `json:"source_url_scheme"`       // 回源协议
	HttpPort        int    `json:"http_port,omitempty"`     // HTTP 端口
	HttpsPort       int    `json:"https_port,omitempty"`    // HTTPS 端口
	IPGroupCode     string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *QiniuCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
