package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/baidubce/bce-sdk-go/bce"
	"github.com/baidubce/bce-sdk-go/services/cdn"
	"github.com/baidubce/bce-sdk-go/services/cdn/api"
)

// BaiduCDNConfig 百度云 CDN 配置
type BaiduCDNConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	SecretAccessKey string `json:"secret_access_key"`  // 密钥Key
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证百度云 CDN 配置
func (cfg *BaiduCDNConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.SecretAccessKey == "" {
		return fmt.Errorf("AccessKeySecret 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultBaiduCDNConfig 创建默认的百度云 CDN 配置
func NewDefaultBaiduCDNConfig() *BaiduCDNConfig {
	return &BaiduCDNConfig{
		Endpoint: "https://cdn.baidubce.com",
	}
}

// ParseCustomConfig 解析 百度云 CDN 自定义配置
func (p *BaiduCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// BaiduCDNProvider 百度云 CDN 提供程序
type BaiduCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *BaiduCDNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *BaiduCDNCustomConfig
	Provider     struct {
		Client *cdn.Client
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData api.DomainConfig
}

// NewBaiduCDNProvider 创建百度云 CDN 提供程序
func NewBaiduCDNProvider(ctx context.Context, logger *slog.Logger, cfg *BaiduCDNConfig, opts ...Option) (provider.CDNProvider, error) {
	return NewBaiduCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewBaiduCDNProviderRaw 创建百度云 CDN 提供程序
func NewBaiduCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *BaiduCDNConfig, opts ...Option) (*BaiduCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &BaiduCDNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &BaiduCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 实例化
	p.Provider.Client, err = cdn.NewClient(
		p.cfg.AccessKeyID,
		p.cfg.SecretAccessKey,
		p.cfg.Endpoint,
	)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Client.HTTPClient = &http.Client{
			Transport: p.opts.httpTransport.Instance(nil),
		}
	}

	// 定义最大重试次数
	p.Provider.Client.Config.Retry = bce.NewBackOffRetryPolicy(3, 20000, 300)

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *BaiduCDNProvider) Name() string {
	return "百度云 CDN"
}

// Check 检查子域名是否存在
// https://console.bce.baidu.com/support/#/api?product=CDN&project=CDN&parent=查询用户名下所有域名
// https://console.bce.baidu.com/support/#/api?product=CDN&project=CDN&parent=查询加速域名详情
func (p *BaiduCDNProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (data *provider.CDNProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// HTTP回源端口，默认80
	if p.CustomConfig.HttpPort == 0 {
		p.CustomConfig.HttpPort = 80
	}

	// HTTPS回源端口，默认443
	if p.CustomConfig.HttpsPort == 0 {
		p.CustomConfig.HttpsPort = 443
	}

	// 源站配置
	originValues := make([]string, 0, len(newOriginValues)*2)
	for _, originValue := range newOriginValues {
		originValues = append(originValues, fmt.Sprintf("http://[%v]:%v", originValue, p.CustomConfig.HttpPort))
		originValues = append(originValues, fmt.Sprintf("https://[%v]:%v", originValue, p.CustomConfig.HttpsPort))
	}

	// 查询用户名下所有域名
	responseCdnList, err := p.Provider.Client.GetDomainStatus("ALL", "")
	if err != nil {
		return nil, fmt.Errorf("请求 GetDomainStatus 失败 %w", err)
	}
	isExist := false
	for _, item := range responseCdnList {
		// 查找域名
		itemFullDomain := item.Domain
		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {
			isExist = true
			break
		}
	}
	// 未找到记录
	if !isExist {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}

	// 查询用户名下所有域名
	responseCdnInfo, err := p.Provider.Client.GetDomainConfig(fullDomain)
	if err != nil {
		return nil, fmt.Errorf("接口 GetDomainConfig 失败 %w", err)
	}
	// 响应参数

	// 查找域名
	// itemFullDomain := responseInfo.Domain

	// 记录数据
	p.ItemData = *responseCdnInfo

	// 查找记录值列表
	itemOriginValues := make([]string, 0, len(responseCdnInfo.Origin))
	for _, originValue := range responseCdnInfo.Origin {
		itemOriginValues = append(itemOriginValues, originValue.Peer)
	}
	// 查找类型
	itemForm := responseCdnInfo.Form
	// 查询协议跟随回源
	itemFollowProtocol, err := p.Provider.Client.GetOriginProtocol(fullDomain)
	if err != nil {
		return nil, fmt.Errorf("接口 GetOriginProtocol 失败 %w", err)
	}

	// 域名ID
	// p.InternalData.RecordID = responseCdnInfo.DomainID

	// CDN域名
	p.InternalData.CdnDomain = utils.TrimDots(responseCdnInfo.Cname)

	// 判断记录值列表
	if !utils.EqualUrlSliceUnordered(itemOriginValues, originValues) {
		p.InternalData.IsContinue = true
		p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", originValues, itemOriginValues)
		return p.InternalData, nil
	}

	// 判断类型
	if !utils.EqualDomainIgnoreCase(itemForm, p.CustomConfig.Form) {
		p.InternalData.IsContinue = true
		p.InternalData.Tips = fmt.Sprintf("类型不匹配，期望 %v，实际 %v", p.CustomConfig.Form, itemForm)
		return p.InternalData, nil
	}

	// 判断是否开启了协议跟随回源
	if !utils.EqualIgnoreCase(itemFollowProtocol, p.CustomConfig.FollowProtocol) {
		p.InternalData.IsContinue = true
		p.InternalData.Tips = fmt.Sprintf("是否开启了协议跟随回源不匹配，期望 %v，实际 %v", p.CustomConfig.FollowProtocol, itemFollowProtocol)
		return p.InternalData, nil
	}

	p.InternalData.IsContinue = false
	return p.InternalData, nil

}

// Update 更新 CDN 记录
// https://cloud.baidu.com/doc/CDN/s/Fk121xrty#创建加速域名接口-createdomain
// https://cloud.baidu.com/doc/CDN/s/Vk121slzt#更新加速域名回源地址-setdomainorigin
// https://cloud.baidu.com/doc/CDN/s/qkggncsxp
func (p *BaiduCDNProvider) Update(ctx context.Context, domain string, subdomain string, newOriginValues []string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// HTTP回源端口，默认80
	if p.CustomConfig.HttpPort == 0 {
		p.CustomConfig.HttpPort = 80
	}

	// HTTPS回源端口，默认443
	if p.CustomConfig.HttpsPort == 0 {
		p.CustomConfig.HttpsPort = 443
	}

	// 源站配置
	sources := make([]api.OriginPeer, 0, len(newOriginValues)*2)
	for _, originValue := range newOriginValues {
		httpSrc := api.OriginPeer{
			Peer:   fmt.Sprintf("http://[%v]:%v", originValue, p.CustomConfig.HttpPort),
			Backup: false,
		}
		httpsSrc := api.OriginPeer{
			Peer:   fmt.Sprintf("https://[%v]:%v", originValue, p.CustomConfig.HttpsPort),
			Backup: false,
		}
		sources = append(sources, httpSrc)
		sources = append(sources, httpsSrc)
	}

	if p.InternalData.CdnDomain == "" {

		// 查询域名是否可添加
		responseCdnValid, err := p.Provider.Client.IsValidDomain(fullDomain)
		if err != nil {
			return fmt.Errorf("接口 IsValidDomain 失败 %w", err)
		}
		// 校验域名是否可添加
		if !responseCdnValid.IsValid {
			return fmt.Errorf("域名 %v 不可添加", fullDomain)
		}

		// 创建加速域名
		requestCdnCreate := &api.OriginInit{}
		requestCdnCreate.Origin = sources           // 源站配置
		requestCdnCreate.DefaultHost = fullDomain   // 域名级别回源host
		requestCdnCreate.Form = p.CustomConfig.Form // 类型
		responseCreate, err := p.Provider.Client.CreateDomain(fullDomain, requestCdnCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateDomain 失败 %w", err)
		}

		// CDN域名
		p.InternalData.CdnDomain = utils.TrimDots(responseCreate.Cname)

		// 设置协议跟随回源
		err = p.Provider.Client.SetOriginProtocol(fullDomain, p.CustomConfig.FollowProtocol)
		if err != nil {
			return fmt.Errorf("接口 SetOriginProtocol 失败 %w", err)
		}

		// 设置IPv6
		err = p.Provider.Client.SetIPv6(fullDomain, true)
		if err != nil {
			return fmt.Errorf("接口 SetIPv6 失败 %w", err)
		}

	} else if p.InternalData.CdnDomain != "" {

		defaultHost := p.ItemData.DefaultHost
		if defaultHost == "" {
			defaultHost = fullDomain
		}

		// 更新加速域名回源地址
		err := p.Provider.Client.SetDomainOrigin(fullDomain, sources, defaultHost)
		if err != nil {
			return fmt.Errorf("接口 SetDomainOrigin 失败 %w", err)
		}

		// 设置协议跟随回源
		err = p.Provider.Client.SetOriginProtocol(fullDomain, p.CustomConfig.FollowProtocol)
		if err != nil {
			return fmt.Errorf("接口 SetOriginProtocol 失败 %w", err)
		}

		// 设置IPv6
		err = p.Provider.Client.SetIPv6(fullDomain, true)
		if err != nil {
			return fmt.Errorf("接口 SetIPv6 失败 %w", err)
		}

	}

	return nil
}

// 百度云 CDN 自定义配置
type BaiduCDNCustomConfig struct {
	Form           string `json:"form"`                    // 类型
	FollowProtocol string `json:"follow_protocol"`         // 是否开启了协议跟随回源
	HttpPort       int    `json:"http_port,omitempty"`     // HTTP 端口
	HttpsPort      int    `json:"https_port,omitempty"`    // HTTPS 端口
	IPGroupCode    string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *BaiduCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
