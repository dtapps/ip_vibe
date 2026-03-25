package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/helper/aliyun_helper"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	cdn20180510 "github.com/alibabacloud-go/cdn-20180510/v9/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"github.com/spf13/cast"
	"go.dtapp.net/library/utils/gojson"
)

// AliyunCDNConfig 阿里云 CDN 配置
type AliyunCDNConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 访问密钥ID
	AccessKeySecret string `json:"access_key_secret"`  // 访问密钥Key
}

// DefaultValidate 验证配置
func (cfg *AliyunCDNConfig) DefaultValidate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("AccessKeySecret 不能为空")
	}
	return nil
}

// NewDefaultAliyunCDNConfig 创建默认的阿里云 CDN 配置
func NewDefaultAliyunCDNConfig() *AliyunCDNConfig {
	return &AliyunCDNConfig{
		Endpoint: "cdn.aliyuncs.com",
	}
}

// ParseCustomConfig 解析 阿里云 CDN 自定义配置
func (p *AliyunCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// AliyunCDNProvider 阿里云 CDN 提供程序
type AliyunCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *AliyunCDNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *AliyunCDNCustomConfig
	Provider     struct {
		Auth    credential.Credential
		Cfg     *openapi.Config
		Client  *cdn20180510.Client
		Runtime *util.RuntimeOptions
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	ItemData cdn20180510.DescribeUserDomainsResponseBodyDomainsPageData
}

// NewAliyunCDNProvider 创建阿里云 CDN 提供程序
func NewAliyunCDNProvider(ctx context.Context, logger *slog.Logger, cfg *AliyunCDNConfig, opts ...Option) (provider.CDNProvider, error) {
	return NewAliyunCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewAliyunCDNProviderRaw 创建阿里云 CDN 提供程序
func NewAliyunCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *AliyunCDNConfig, opts ...Option) (*AliyunCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &AliyunCDNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &AliyunCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth, err = credential.NewCredential(&credential.Config{
		Type:            new("access_key"),
		AccessKeyId:     new(p.cfg.AccessKeyID),
		AccessKeySecret: new(p.cfg.AccessKeySecret),
	})
	if err != nil {
		return nil, fmt.Errorf("创建凭证失败 %w", err)
	}

	// 选项
	p.Provider.Cfg = &openapi.Config{
		Credential: p.Provider.Auth,
		Endpoint:   new(p.cfg.Endpoint),
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = cdnHttpTransport
	}
	if p.opts.httpTransport != nil {
		httpHandler := aliyun_helper.NewHttpHandler(p.opts.httpTransport)
		p.Provider.Cfg.HttpClient = httpHandler
	}

	// 实例化
	p.Provider.Client, err = cdn20180510.NewClient(p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 定义最大重试次数
	p.Provider.Runtime = &dara.RuntimeOptions{
		Autoretry:   new(true),
		MaxAttempts: new(3),
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *AliyunCDNProvider) Name() string {
	return "阿里云 CDN"
}

// Check 检查子域名是否存在
// https://next.api.aliyun.com/api/ESA/2024-09-10/ListRecords
func (p *AliyunCDNProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (data *provider.CDNProviderInternalData, err error) {

	defaultIP := "192.0.2.0"

	// 查询加速域名
	requestCdnQuery := &cdn20180510.DescribeUserDomainsRequest{}
	requestCdnQuery.DomainName = new(full_domain) // 加速域名
	requestCdnQuery.PageNumber = tea.Int32(1)     // 取得第几页
	requestCdnQuery.PageSize = tea.Int32(500)     // 分页大小
	responseCdnQuery, err := p.Provider.Client.DescribeUserDomainsWithContext(ctx, requestCdnQuery, p.Provider.Runtime)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeUserDomainsWithContext 失败 %w", err)
	}
	// 响应参数
	if responseCdnQuery.GetBody().GetDomains().GetPageData() == nil || len(responseCdnQuery.GetBody().GetDomains().GetPageData()) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseCdnQuery.GetBody().GetDomains().GetPageData() {

		// 查找域名
		itemFullDomain := tea.StringValue(item.GetDomainName())

		if utils.EqualDomainIgnoreCase(itemFullDomain, full_domain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录值列表
			itemOriginValues := make([]string, 0, len(item.GetSources().GetSource()))
			for _, source := range item.GetSources().GetSource() {
				if tea.StringValue(source.GetContent()) == defaultIP {
					continue // 跳过默认测试地址
				}
				itemOriginValues = append(itemOriginValues, tea.StringValue(source.GetContent()))
			}
			// 查找加速域名的业务类型
			itemCdnType := tea.StringValue(item.GetCdnType())

			// 域名ID
			p.InternalData.RecordID = cast.ToString(tea.Int64Value(item.GetDomainId()))

			// CDN域名
			p.InternalData.CdnDomain = utils.TrimDotsPrt(item.GetCname())

			// 判断记录值列表
			if !utils.EqualStringSliceUnordered(itemOriginValues, origin_values) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值列表不匹配，期望 %v，实际 %v", origin_values, itemOriginValues)
				return p.InternalData, nil
			}

			// 判断业务类型
			if !utils.EqualDomainIgnoreCase(itemCdnType, p.CustomConfig.CdnType) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("业务类型不匹配，期望 %v，实际 %v", p.CustomConfig.CdnType, itemCdnType)
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
// https://next.api.aliyun.com/api/Cdn/2018-05-10/AddCdnDomain
// https://next.api.aliyun.com/api/Cdn/2018-05-10/ModifyCdnDomain
func (p *AliyunCDNProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (err error) {

	type SourceItem struct {
		Content  string `json:"content"`  // 回源地址
		Type     string `json:"type"`     // 回源地址类型
		Priority string `json:"priority"` // 优先级
		Port     int    `json:"port"`     // 端口
		Weight   string `json:"weight"`   // 权重
	}
	sources := make([]SourceItem, 0, len(origin_values)*2)

	// 计算平均权重
	n := len(origin_values)
	weights := make([]int, n)
	if n > 0 {
		base := 100 / n
		remainder := 100 % n
		for i := range n {
			weights[i] = base
			if i < remainder {
				weights[i]++ // 前 remainder 个加 1，保证总和=100
			}
		}
	}

	// TODO: 默认加一个addr
	defaultIP := "192.0.2.0"
	origin_values = append(origin_values, defaultIP)

	for i, addr := range origin_values {
		// 判断是否是默认测试地址
		isDefaultTestIP := (addr == defaultIP)

		// 构建 HTTP 源站（如果端口启用）
		if p.CustomConfig.HttpPort > 0 {
			src := SourceItem{
				Content: addr,
				Type:    "ipaddr",
				Port:    p.CustomConfig.HttpPort,
			}
			if !isDefaultTestIP {
				src.Priority = "20"
				src.Weight = cast.ToString(weights[i])
			}
			sources = append(sources, src)
		}

		// 构建 HTTPS 源站（如果端口启用）
		if p.CustomConfig.HttpsPort > 0 {
			src := SourceItem{
				Content: addr,
				Type:    "ipaddr",
				Port:    p.CustomConfig.HttpsPort,
			}
			if !isDefaultTestIP {
				src.Priority = "20"
				src.Weight = cast.ToString(weights[i])
			}
			sources = append(sources, src)
		}
	}

	if p.InternalData.RecordID == "" {

		// 创建加速域名
		requestCdnCreate := &cdn20180510.AddCdnDomainRequest{}
		requestCdnCreate.CdnType = new(p.CustomConfig.CdnType)            // 加速域名的业务类型
		requestCdnCreate.DomainName = new(full_domain)                    // 需要接入CDN的加速域名
		requestCdnCreate.Sources = new(gojson.JsonEncodeNoError(sources)) // 回源地址列表
		requestCdnCreate.Scope = new(p.CustomConfig.Scope)                // 加速区域
		_, err = p.Provider.Client.AddCdnDomainWithContext(ctx, requestCdnCreate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 AddCdnDomainWithContext 失败 %w", err)
		}

	} else if p.InternalData.RecordID != "" {

		// 更新加速域名
		requestCdnUpdate := &cdn20180510.ModifyCdnDomainRequest{}
		requestCdnUpdate.DomainName = new(full_domain)                    // 需要接入CDN的加速域名
		requestCdnUpdate.Sources = new(gojson.JsonEncodeNoError(sources)) // 回源地址列表
		_, err = p.Provider.Client.ModifyCdnDomainWithContext(ctx, requestCdnUpdate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 ModifyCdnDomainWithContext 失败 %w", err)
		}

	}

	return nil
}

// 阿里云 CDN 自定义配置
type AliyunCDNCustomConfig struct {
	CdnType       string `json:"cdn_type"`                   // 加速域名的业务类型
	Scope         string `json:"scope,omitempty"`            // 加速区域
	HttpPort      int    `json:"http_port,omitempty"`        // HTTP 端口
	HttpsPort     int    `json:"https_port,omitempty"`       // HTTPS 端口
	IPGroupIDCode string `json:"ip_group_id_code,omitempty"` // 设备IP组代码
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *AliyunCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
