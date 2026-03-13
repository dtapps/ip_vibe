package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

// TencentDNSConfig 腾讯云 DNS 配置
type TencentDNSConfig struct {
	Endpoint    string `json:"endpoint,omitempty"`     // 接口地址
	SecretID    string `json:"secret_id"`              // 密钥ID
	SecretKey   string `json:"secret_key"`             // 密钥Key
	DomainGrade string `json:"domain_grade,omitempty"` // 域名套餐等级
}

// DefaultValidate 验证配置
func (cfg *TencentDNSConfig) DefaultValidate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	return nil
}

// NewDefaultTencentDNSConfig 创建默认的腾讯云 DNS 配置
func NewDefaultTencentDNSConfig() *TencentDNSConfig {
	return &TencentDNSConfig{
		Endpoint: "dnspod.tencentcloudapi.com",
	}
}

// ParseCustomConfig 解析 腾讯云 DNS 自定义配置
func (p *TencentDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// TencentDNSProvider 腾讯云 DNS 提供程序
type TencentDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *TencentDNSConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *TencentDNSCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *dnspod.Client
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData dnspod.RecordListItem
}

// NewTencentDNSProvider 创建腾讯云 DNS 提供程序
func NewTencentDNSProvider(ctx context.Context, logger *slog.Logger, cfg *TencentDNSConfig, opts ...Option) (provider.DNSProvider, error) {
	return NewTencentDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewTencentDNSProviderRaw 创建腾讯云 DNS 提供程序
func NewTencentDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *TencentDNSConfig, opts ...Option) (*TencentDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentDNSProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &TencentDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
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

	// 实例化client选项
	p.Provider.Cfg = profile.NewClientProfile()
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint // 指定域名
	p.Provider.Cfg.NetworkFailureMaxRetries = 3          // 定义最大重试次数

	// 开启日志
	p.Provider.Cfg.Debug = sys_env.GvaEnv.DEBUG

	// 实例化
	p.Provider.Client, err = dnspod.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, fmt.Errorf("创建客户端失败 %w", err)
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport == nil {
		p.opts.httpTransport = dnsHttpTransport
	}
	if p.opts.httpTransport != nil {
		p.Provider.Client.WithHttpTransport(p.opts.httpTransport.Instance(nil))
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *TencentDNSProvider) Name() string {
	return "腾讯云 DNS"
}

// Check 检查子域名是否存在
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordList
func (p *TencentDNSProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (data *provider.DNSProviderInternalData, err error) {

	// 查询域名记录
	requestDnsQuery := dnspod.NewDescribeRecordListRequest()
	requestDnsQuery.Domain = common.StringPtr(domain_name)   // 域名
	requestDnsQuery.Subdomain = common.StringPtr(sub_domain) // 解析记录的主机头
	if record_type != "" {
		requestDnsQuery.RecordType = common.StringPtr(record_type) // 获取某种类型的解析记录
	}
	responseDnsQuery, err := p.Provider.Client.DescribeRecordListWithContext(ctx, requestDnsQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeRecordListWithContext 失败 %w", err)
	}
	// 响应参数
	if responseDnsQuery.Response.RecordList == nil || len(responseDnsQuery.Response.RecordList) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseDnsQuery.Response.RecordList {

		if utils.EqualDomainIgnoreCase(utils.PtrToString(item.Name), sub_domain) {

			// 记录数据
			p.ItemData = *item

			// 查找记录类型
			itemRecordType := utils.PtrToString(item.Type)
			// 查找记录值
			itemOriginValue := utils.PtrToString(item.Value)
			// 查找记录线路
			itemRecordLine := utils.PtrToString(item.Line)
			// 查找备注
			itemRemark := utils.PtrToString(item.Remark)

			// 检查记录类型
			if record_type != "" && !utils.EqualDomainIgnoreCase(itemRecordType, record_type) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = cast.ToString(utils.PtrToUint64(item.RecordId))

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, origin_value) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", origin_value, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录对应的线路
			if !utils.EqualDomainIgnoreCase(itemRecordLine, p.CustomConfig.RecordLineValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录对应的线路不匹配，期望 %v，实际 %v", p.CustomConfig.RecordLineValue, itemRecordLine)
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
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=CreateRecord
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=ModifyRecord
func (p *TencentDNSProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (err error) {

	if p.InternalData.RecordID == "" {

		// 创建域名记录
		requestDnsCreate := dnspod.NewCreateRecordRequest()
		requestDnsCreate.Domain = common.StringPtr(domain_name)                        // 域名
		requestDnsCreate.RecordType = common.StringPtr(record_type)                    // 记录类型
		requestDnsCreate.Value = common.StringPtr(origin_value)                        // 记录值
		requestDnsCreate.SubDomain = common.StringPtr(sub_domain)                      // 主机记录
		requestDnsCreate.RecordLine = common.StringPtr(p.CustomConfig.RecordLineValue) // 记录线路
		requestDnsCreate.Remark = common.StringPtr(p.CustomConfig.Remark)              // 备注
		responseDnsCreate, err := p.Provider.Client.CreateRecordWithContext(ctx, requestDnsCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateRecordWithContext 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = cast.ToString(utils.PtrToUint64(responseDnsCreate.Response.RecordId))

	} else if p.InternalData.RecordID != "" {

		// 更新域名记录
		requestDnsUpdate := dnspod.NewModifyRecordRequest()
		requestDnsUpdate.Domain = common.StringPtr(domain_name)                              // 域名
		requestDnsUpdate.RecordType = common.StringPtr(record_type)                          // 记录类型
		requestDnsUpdate.Value = common.StringPtr(origin_value)                              // 记录值
		requestDnsUpdate.RecordId = common.Uint64Ptr(cast.ToUint64(p.InternalData.RecordID)) // 记录 ID
		requestDnsUpdate.SubDomain = common.StringPtr(sub_domain)                            // 主机记录
		requestDnsUpdate.RecordLine = common.StringPtr(p.CustomConfig.RecordLineValue)       // 记录线路
		requestDnsUpdate.Remark = common.StringPtr(p.CustomConfig.Remark)                    // 备注
		responseDnsUpdate, err := p.Provider.Client.ModifyRecordWithContext(ctx, requestDnsUpdate)
		if err != nil {
			return fmt.Errorf("接口 ModifyRecordWithContext 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = cast.ToString(utils.PtrToUint64(responseDnsUpdate.Response.RecordId))

	}

	return nil
}

// 腾讯云 DNS 自定义配置
type TencentDNSCustomConfig struct {
	DomainText      string `json:"domain_text"`                // 域名
	DomainValue     string `json:"domain_value"`               // 域名编号
	RecordLineValue string `json:"record_line_value"`          // 线路
	IPGroupIDCode   string `json:"ip_group_id_code,omitempty"` // 设备IP组代码
	Remark          string `json:"remark,omitempty"`           // 备注
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *TencentDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// TencentDNSProviderLine 腾讯云 DNS 线路
type TencentDNSProviderLine struct {
	LineID   string `json:"line_id"`   // 线路ID
	LineName string `json:"line_name"` // 线路名称
}

// 获取腾讯云 DNS 线路列表
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordLineList
func (p *TencentDNSProvider) ListLines(ctx context.Context) (lines []*TencentDNSProviderLine, err error) {
	// 判断是否配置了域名等级
	if p.cfg.DomainGrade == "" {
		return nil, nil
	}
	// 查询线路列表
	requestQueryLine := dnspod.NewDescribeRecordLineListRequest()
	requestQueryLine.Domain = common.StringPtr(p.CustomConfig.DomainText)
	requestQueryLine.DomainGrade = common.StringPtr(p.cfg.DomainGrade)
	responseQueryLine, err := p.Provider.Client.DescribeRecordLineListWithContext(ctx, requestQueryLine)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeRecordLineListWithContext 失败 %w", err)
	}
	for _, item := range responseQueryLine.Response.LineList {
		lines = append(lines, &TencentDNSProviderLine{
			LineID:   utils.PtrToString(item.Name),
			LineName: utils.PtrToString(item.Name),
		})
	}
	return lines, nil
}

// 获取域名列表
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeDomainList
func (p *TencentDNSProvider) DescribeDomainList(ctx context.Context) ([]*dnspod.DomainListItem, error) {
	request := dnspod.NewDescribeDomainListRequest()
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(3000)
	response, err := p.Provider.Client.DescribeDomainListWithContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeDomainListWithContext 失败 %w", err)
	}
	return response.Response.DomainList, nil
}
