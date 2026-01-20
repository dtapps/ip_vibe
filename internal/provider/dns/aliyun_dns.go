package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/helper/aliyun_helper"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/dara"
	"github.com/alibabacloud-go/tea/tea"
	credential "github.com/aliyun/credentials-go/credentials"
	"go.dtapp.net/library/utils/gojson"
)

// AliyunDNSConfig 阿里云 DNS 配置
type AliyunDNSConfig struct {
	Endpoint        string `json:"endpoint,omitempty"` // 接口地址
	AccessKeyID     string `json:"access_key_id"`      // 密钥ID
	AccessKeySecret string `json:"access_key_secret"`  // 密钥Key
	DomainName      string `json:"domain_name"`        // 主域名
}

// Validate 验证阿里云 DNS 配置
func (cfg *AliyunDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.AccessKeyID == "" {
		return fmt.Errorf("AccessKeyID 不能为空")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("AccessKeySecret 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultAliyunDNSConfig 创建默认的阿里云 DNS 配置
func NewDefaultAliyunDNSConfig() *AliyunDNSConfig {
	return &AliyunDNSConfig{
		Endpoint: "alidns.aliyuncs.com",
	}
}

// ParseCustomConfig 解析 阿里云 DNS 自定义配置
func (p *AliyunDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// AliyunDNSProvider 阿里云 DNS 提供程序
type AliyunDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *AliyunDNSConfig
	opts   *dnsOptions
	logger *slog.Logger

	CustomConfig *AliyunDNSCustomConfig
	Provider     struct {
		Auth    credential.Credential
		Cfg     *openapi.Config
		Client  *alidns20150109.Client
		Runtime *util.RuntimeOptions
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData alidns20150109.DescribeDomainRecordsResponseBodyDomainRecordsRecord
}

// NewAliyunDNSProvider 创建阿里云 DNS 提供程序
func NewAliyunDNSProvider(ctx context.Context, logger *slog.Logger, cfg *AliyunDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewAliyunDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewAliyunDNSProviderRaw 创建阿里云 DNS 提供程序
func NewAliyunDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *AliyunDNSConfig, opts ...DnsOption) (*AliyunDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &AliyunDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		logger:       logger,
		CustomConfig: &AliyunDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth, err = credential.NewCredential(&credential.Config{
		Type:            tea.String("access_key"),
		AccessKeyId:     tea.String(p.cfg.AccessKeyID),
		AccessKeySecret: tea.String(p.cfg.AccessKeySecret),
	})
	if err != nil {
		return nil, fmt.Errorf("创建凭证失败 %w", err)
	}

	// 选项
	p.Provider.Cfg = &openapi.Config{
		Credential: p.Provider.Auth,
		Endpoint:   tea.String(p.cfg.Endpoint),
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		httpHandler := aliyun_helper.NewHttpHandler(p.opts.httpTransport)
		p.Provider.Cfg.HttpClient = httpHandler
	}

	// 实例化
	p.Provider.Client, err = alidns20150109.NewClient(p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 运行时
	p.Provider.Runtime = &dara.RuntimeOptions{
		Autoretry:   tea.Bool(true),
		MaxAttempts: tea.Int(3),
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *AliyunDNSProvider) Name() string {
	return "阿里云 DNS"
}

// printData 打印数据
func (p *AliyunDNSProvider) printData(ctx context.Context) {
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
// https://next.api.aliyun.com/api/Alidns/2015-01-09/DescribeDomainRecords
func (p *AliyunDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

	// 查询域名记录
	requestDnsQuery := &alidns20150109.DescribeDomainRecordsRequest{}
	requestDnsQuery.SearchMode = tea.String("COMBINATION") // 搜索模式
	requestDnsQuery.DomainName = tea.String(domain)        // 域名名称
	requestDnsQuery.RRKeyWord = tea.String(subdomain)      // 主机记录的关键字
	if recordType != "" {
		requestDnsQuery.TypeKeyWord = tea.String(recordType) // 解析类型的关键字
	}
	responseDnsQuery, err := p.Provider.Client.DescribeDomainRecordsWithContext(ctx, requestDnsQuery, p.Provider.Runtime)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeDomainRecordsWithContext 失败 %w", err)
	}
	// 响应参数
	if responseDnsQuery.GetBody().GetDomainRecords().GetRecord() == nil || len(responseDnsQuery.GetBody().GetDomainRecords().GetRecord()) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseDnsQuery.GetBody().GetDomainRecords().GetRecord() {

		if utils.EqualDomainIgnoreCase(tea.StringValue(item.GetRR()), subdomain) && utils.EqualDomainIgnoreCase(tea.StringValue(item.GetDomainName()), domain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录类型
			itemRecordType := tea.StringValue(item.GetType())
			// 查找记录值
			itemOriginValue := tea.StringValue(item.GetValue())
			// 查找解析记录的线路类型
			itemLine := tea.StringValue(item.GetLine())

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = tea.StringValue(item.GetRecordId())

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", newOriginValue, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录对应的线路
			if !utils.EqualDomainIgnoreCase(itemLine, p.CustomConfig.Line) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录对应的线路不匹配，期望 %v，实际 %v", p.CustomConfig.Line, itemLine)
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
// https://next.api.aliyun.com/api/Alidns/2015-01-09/AddDomainRecord
// https://next.api.aliyun.com/api/Alidns/2015-01-09/UpdateDomainRecord
func (p *AliyunDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 打印数据
	p.printData(ctx)

	if p.InternalData.RecordID == "" {

		// 创建域名记录
		requestDnsCreate := &alidns20150109.AddDomainRecordRequest{}
		requestDnsCreate.DomainName = tea.String(domain)        // 域名名称
		requestDnsCreate.RR = tea.String(subdomain)             // 主机记录
		requestDnsCreate.Type = tea.String(recordType)          // 解析记录类型
		requestDnsCreate.Value = tea.String(originValue)        // 记录值
		requestDnsCreate.Line = tea.String(p.CustomConfig.Line) // 解析线路
		responseDnsCreate, err := p.Provider.Client.AddDomainRecordWithContext(ctx, requestDnsCreate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 AddDomainRecordWithContext 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = tea.StringValue(responseDnsCreate.GetBody().GetRecordId())

	} else if p.InternalData.RecordID != "" {

		// 更新域名记录
		requestDnsUpdate := &alidns20150109.UpdateDomainRecordRequest{}
		requestDnsUpdate.RecordId = tea.String(p.InternalData.RecordID) // 解析记录的ID
		requestDnsUpdate.RR = tea.String(subdomain)                     // 主机记录
		requestDnsUpdate.Type = tea.String(recordType)                  // 解析记录类型
		requestDnsUpdate.Value = tea.String(originValue)                // 记录值
		requestDnsUpdate.Line = tea.String(p.CustomConfig.Line)         // 解析线路
		responseDnsUpdate, err := p.Provider.Client.UpdateDomainRecordWithContext(ctx, requestDnsUpdate, p.Provider.Runtime)
		if err != nil {
			return fmt.Errorf("接口 UpdateDomainRecordWithContext 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = tea.StringValue(responseDnsUpdate.GetBody().GetRecordId())

	}

	return nil
}

// 阿里云 DNS 自定义配置
type AliyunDNSCustomConfig struct {
	Line        string `json:"line"`                    // 解析记录的线路类型
	IPGroupCode string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *AliyunDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// AliyunDNSProviderLine 阿里云 DNS 线路
type AliyunDNSProviderLine struct {
	LineID   string `json:"line_id"`   // 线路ID
	LineName string `json:"line_name"` // 线路名称
}

// 获取阿里云 DNS 线路列表
// https://next.api.aliyun.com/api/Alidns/2015-01-09/DescribeSupportLines
func (p *AliyunDNSProvider) ListLines(ctx context.Context) (lines []*AliyunDNSProviderLine, err error) {
	// 查询线路列表
	requestQueryLine := &alidns20150109.DescribeSupportLinesRequest{}
	requestQueryLine.DomainName = tea.String(p.cfg.DomainName) // 域名名称
	responseQueryLine, err := p.Provider.Client.DescribeSupportLinesWithContext(ctx, requestQueryLine, p.Provider.Runtime)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeSupportLinesWithContext 失败 %w", err)
	}
	for _, item := range responseQueryLine.GetBody().GetRecordLines().GetRecordLine() {
		lines = append(lines, &AliyunDNSProviderLine{
			LineID:   tea.StringValue(item.GetLineCode()),
			LineName: tea.StringValue(item.GetLineDisplayName()),
		})
	}
	return lines, nil
}
