package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"go.dtapp.net/library/utils/gojson"
)

// TencentDNSConfig 腾讯云 DNS 配置
type TencentDNSConfig struct {
	Endpoint    string `json:"endpoint,omitempty"`     // 接口地址
	SecretID    string `json:"secret_id"`              // 密钥ID
	SecretKey   string `json:"secret_key"`             // 密钥Key
	DomainName  string `json:"domain_name"`            // 主域名
	DomainGrade string `json:"domain_grade,omitempty"` // 域名套餐等级
}

// Validate 验证腾讯云 DNS 配置
func (cfg *TencentDNSConfig) Validate() error {
	if cfg.Endpoint == "" {
		return fmt.Errorf("Endpoint 不能为空")
	}
	if cfg.SecretID == "" {
		return fmt.Errorf("SecretID 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
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
		return err
	}
	return nil
}

// TencentDNSProvider 腾讯云 DNS 提供程序
type TencentDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *TencentDNSConfig
	opts   *dnsOptions
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
func NewTencentDNSProvider(ctx context.Context, logger *slog.Logger, cfg *TencentDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewTencentDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewTencentDNSProviderRaw 创建腾讯云 DNS 提供程序
func NewTencentDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *TencentDNSConfig, opts ...DnsOption) (*TencentDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &TencentDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
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
	p.Provider.Cfg.HttpProfile.Endpoint = p.cfg.Endpoint

	// 实例化
	p.Provider.Client, err = dnspod.NewClient(p.Provider.Auth, "", p.Provider.Cfg)
	if err != nil {
		return nil, err
	}

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
		p.Provider.Client.WithHttpTransport(p.opts.httpTransport.Instance(nil))
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *TencentDNSProvider) Name() string {
	return "腾讯云 DNS"
}

// printData 打印数据
func (p *TencentDNSProvider) printData(ctx context.Context) {
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
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordList
func (p *TencentDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

	// 请求参数
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)       // 域名
	request.Subdomain = common.StringPtr(subdomain) // 解析记录的主机头
	if recordType != "" {
		request.RecordType = common.StringPtr(recordType) // 获取某种类型的解析记录
	}
	// 请求
	response, err := p.Provider.Client.DescribeRecordListWithContext(ctx, request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	if response.Response.RecordList == nil || len(response.Response.RecordList) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range response.Response.RecordList {

		if utils.EqualDomainIgnoreCase(utils.PtrToString(item.Name), subdomain) {

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
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = cast.ToString(*item.RecordId)

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", newOriginValue, itemOriginValue)
				return p.InternalData, nil
			}

			// 判断解析记录对应的线路
			if !utils.EqualDomainIgnoreCase(itemRecordLine, p.CustomConfig.RecordLine) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("解析记录对应的线路不匹配，期望 %v，实际 %v", p.CustomConfig.RecordLine, itemRecordLine)
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
func (p *TencentDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 打印数据
	p.printData(ctx)

	if p.InternalData.RecordID == "" {
		// 请求参数
		request := dnspod.NewCreateRecordRequest()
		request.Domain = common.StringPtr(domain)                        // 域名
		request.RecordType = common.StringPtr(recordType)                // 记录类型
		request.Value = common.StringPtr(originValue)                    // 记录值
		request.SubDomain = common.StringPtr(subdomain)                  // 主机记录
		request.RecordLine = common.StringPtr(p.CustomConfig.RecordLine) // 记录线路
		request.Remark = common.StringPtr(p.CustomConfig.Remark)         // 备注
		// 请求
		response, err := p.Provider.Client.CreateRecordWithContext(ctx, request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if p.InternalData.RecordID != "" {
		// 请求参数
		request := dnspod.NewModifyRecordRequest()
		request.Domain = common.StringPtr(domain)                                   // 域名
		request.RecordType = common.StringPtr(recordType)                           // 记录类型
		request.Value = common.StringPtr(originValue)                               // 记录值
		request.RecordId = common.Uint64Ptr(cast.ToUint64(p.InternalData.RecordID)) // 记录 ID
		request.SubDomain = common.StringPtr(subdomain)                             // 主机记录
		request.RecordLine = common.StringPtr(p.CustomConfig.RecordLine)            // 记录线路
		request.Remark = common.StringPtr(p.CustomConfig.Remark)                    // 备注
		// 请求
		response, err := p.Provider.Client.ModifyRecordWithContext(ctx, request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// 腾讯云 DNS 自定义配置
type TencentDNSCustomConfig struct {
	RecordLine string `json:"record_line"`      // 记录线路
	Remark     string `json:"remark,omitempty"` // 备注
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *TencentDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return err
	}
	return nil
}

// TencentDNSLine 腾讯云 DNS 线路
type TencentDNSLine struct {
	LineID   string `json:"line_id"`   // 线路ID
	LineName string `json:"line_name"` // 线路名称
}

// 获取腾讯云 DNS 线路列表
// https://console.cloud.tencent.com/api/explorer?Product=dnspod&Version=2021-03-23&Action=DescribeRecordLineList
func (p *TencentDNSProvider) ListLines(ctx context.Context) (lines []*TencentDNSLine, err error) {
	// 判断是否配置了域名等级
	if p.cfg.DomainGrade == "" {
		return nil, nil
	}
	// 请求参数
	request := dnspod.NewDescribeRecordLineListRequest()
	request.Domain = common.StringPtr(p.cfg.DomainName)
	request.DomainGrade = common.StringPtr(p.cfg.DomainGrade)
	// 请求
	response, err := p.Provider.Client.DescribeRecordLineListWithContext(ctx, request)
	printJSON("查询线路", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Response.LineList {
		lines = append(lines, &TencentDNSLine{
			LineID:   utils.PtrToString(item.Name),
			LineName: utils.PtrToString(item.Name),
		})
	}
	return
}
