package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/volcengine/volcengine-go-sdk/service/dns"
	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/session"
)

// VolcengineDNSConfig 火山引擎 DNS 配置
type VolcengineDNSConfig struct {
	Region     string `json:"region,omitempty"` // 区域
	AccessKey  string `json:"access_key"`       // 密钥ID
	SecretKey  string `json:"secret_key"`       // 密钥密钥
	ZoneID     int64  `json:"zone_id"`          // 域名ID
	DomainName string `json:"domain_name"`      // 主域名
}

// Validate 验证火山引擎 DNS 配置
func (cfg *VolcengineDNSConfig) Validate() error {
	if cfg.Region == "" {
		return fmt.Errorf("Region 不能为空")
	}
	if cfg.AccessKey == "" {
		return fmt.Errorf("AccessKey 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.ZoneID == 0 {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultVolcengineDNSConfig 创建默认的火山引擎 DNS 配置
func NewDefaultVolcengineDNSConfig() *VolcengineDNSConfig {
	return &VolcengineDNSConfig{
		Region: "cn-beijing",
	}
}

// ParseCustomConfig 解析 火山引擎 DNS 自定义配置
func (p *VolcengineDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// VolcengineDNSProvider 火山引擎 DNS 提供程序
type VolcengineDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg  *VolcengineDNSConfig
	opts *dnsOptions

	CustomConfig *VolcengineDNSCustomConfig
	Provider     struct {
		Auth   *credentials.Credentials
		Cfg    *volcengine.Config
		Sess   *session.Session
		Client *dns.DNS
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData dns.RecordForListRecordsOutput
}

// NewVolcengineDNSProvider 创建火山引擎 DNS 提供程序
func NewVolcengineDNSProvider(ctx context.Context, cfg *VolcengineDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewVolcengineDNSProviderRaw(ctx, cfg, opts...)
}

// NewVolcengineDNSProviderRaw 创建火山引擎 DNS 提供程序
func NewVolcengineDNSProviderRaw(ctx context.Context, cfg *VolcengineDNSConfig, opts ...DnsOption) (*VolcengineDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &VolcengineDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		CustomConfig: &VolcengineDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 认证
	p.Provider.Auth = credentials.NewStaticCredentials(p.cfg.AccessKey, p.cfg.SecretKey, "")

	// 配置
	p.Provider.Cfg = volcengine.NewConfig().
		WithRegion(p.cfg.Region).
		WithCredentials(p.Provider.Auth)

	// 会话
	p.Provider.Sess, err = session.NewSession(p.Provider.Cfg)
	if err != nil {
		return nil, err
	}

	// 实例化
	p.Provider.Client = dns.New(p.Provider.Sess)

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		// 配置
		p.Provider.Cfg = volcengine.NewConfig().
			WithRegion(p.cfg.Region).
			WithCredentials(p.Provider.Auth).
			WithHTTPClient(&http.Client{
				Transport: p.opts.httpTransport.Instance(nil),
			})

		// 会话
		p.Provider.Sess, err = session.NewSession(p.Provider.Cfg)
		if err != nil {
			return nil, err
		}

		// 实例化
		p.Provider.Client = dns.New(p.Provider.Sess)

	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *VolcengineDNSProvider) Name() string {
	return "火山引擎 DNS"
}

// Check 检查子域名是否存在
// https://api.volcengine.com/api-explorer?action=ListRecords
func (p *VolcengineDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := &dns.ListRecordsInput{}
	request.Host = volcengine.String(fullDomain)    // 主机记录
	request.PageNumber = volcengine.Int32(1)        // 页码
	request.PageSize = volcengine.Int32(500)        // 每页包含的解析记录的数量
	request.SearchMode = volcengine.String("exact") // Host 匹配模式
	if recordType != "" {
		request.Type = volcengine.String(recordType) // 记录类型
	}
	request.ZID = volcengine.Int64(p.cfg.ZoneID) // 域名ID
	// 请求
	response, err := p.Provider.Client.ListRecordsWithContext(ctx, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Records == nil || len(response.Records) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range response.Records {

		// 查找域名
		itemFullDomain := volcengine.StringValue(item.Host)
		// 查找记录类型

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = *item

			itemRecordType := utils.PtrToString(item.Type)
			// 查找记录值
			itemOriginValue := utils.PtrToString(item.Value)
			// 查找解析记录对应的线路
			itemLine := utils.PtrToString(item.Line)
			// 查找备注
			itemRemark := utils.PtrToString(item.Remark)

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = volcengine.StringValue(item.RecordID)

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

			// 判断备注
			if !utils.EqualDomainIgnoreCase(itemRemark, p.CustomConfig.Remark) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("备注不匹配，期望 %v，实际 %v", p.CustomConfig.Remark, itemRemark)
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
// https://api.volcengine.com/api-explorer?action=CreateRecord
// https://api.volcengine.com/api-explorer?action=UpdateRecord
func (p *VolcengineDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if p.InternalData.RecordID == "" {
		// 请求参数
		request := &dns.CreateRecordInput{}
		request.Host = volcengine.String(fullDomain)              // 主机记录
		request.Type = volcengine.String(recordType)              // 记录类型
		request.Value = volcengine.String(originValue)            // 解析记录的记录值
		request.ZID = volcengine.Int64(p.cfg.ZoneID)              // 域名ID
		request.Line = volcengine.String(p.CustomConfig.Line)     // 解析记录对应的线路
		request.Remark = volcengine.String(p.CustomConfig.Remark) // 解析记录的备注
		// 请求
		response, err := p.Provider.Client.CreateRecordWithContext(ctx, request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if p.InternalData.RecordID != "" {
		// 请求参数
		request := &dns.UpdateRecordInput{}
		request.Host = volcengine.String(fullDomain)                  // 主机记录
		request.RecordID = volcengine.String(p.InternalData.RecordID) // 需要更新的解析记录的 ID
		request.Type = volcengine.String(recordType)                  // 记录类型
		request.Value = volcengine.String(originValue)                // 解析记录的记录值
		request.Line = volcengine.String(p.CustomConfig.Line)         // 解析记录对应的线路
		request.Remark = volcengine.String(p.CustomConfig.Remark)     // 解析记录的备注
		// 请求
		response, err := p.Provider.Client.UpdateRecordWithContext(ctx, request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 火山引擎 DNS 自定义配置
type VolcengineDNSCustomConfig struct {
	Line   string `json:"line"`             // 解析记录对应的线路
	Remark string `json:"remark,omitempty"` // 备注
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *VolcengineDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}

// VolcengineDNSLine 火山引擎 DNS 线路
type VolcengineDNSLine struct {
	LineID   string `json:"line_id"`   // 线路ID
	LineName string `json:"line_name"` // 线路名称
}

// 获取火山引擎 DNS 线路列表
// https://api.volcengine.com/api-explorer?action=ListLines
func (p *VolcengineDNSProvider) ListLines(ctx context.Context) (lines []*VolcengineDNSLine, err error) {
	// 请求参数
	request := &dns.ListLinesInput{}
	request.ZID = volcengine.Int64(p.cfg.ZoneID) // 域名ID
	// 请求
	response, err := p.Provider.Client.ListLinesWithContext(ctx, request)
	printJSON("查询线路", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Lines {
		lines = append(lines, &VolcengineDNSLine{
			LineID:   volcengine.StringValue(item.Value),
			LineName: volcengine.StringValue(item.Name),
		})
	}
	return
}
