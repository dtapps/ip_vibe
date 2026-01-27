package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/provider"
	"vibe_ip/internal/utils"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/client"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/domainservice/models"
	"github.com/spf13/cast"
)

// JdcloudDNSConfig 京东云 DNS 配置
type JdcloudDNSConfig struct {
	Region     string `json:"region,omitempty"` // 区域
	AccessKey  string `json:"access_key"`       // 密钥ID
	SecretKey  string `json:"secret_key"`       // 密钥密钥
	ZoneID     string `json:"zone_id"`          // 域名ID
	DomainName string `json:"domain_name"`      // 主域名
}

// Validate 验证京东云 DNS 配置
func (cfg *JdcloudDNSConfig) Validate() error {
	if cfg.Region == "" {
		return fmt.Errorf("Region 不能为空")
	}
	if cfg.AccessKey == "" {
		return fmt.Errorf("AccessKey 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	if cfg.ZoneID == "" {
		return fmt.Errorf("ZoneID 不能为空")
	}
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultJdcloudDNSConfig 创建默认的京东云 DNS 配置
func NewDefaultJdcloudDNSConfig() *JdcloudDNSConfig {
	return &JdcloudDNSConfig{
		Region: "cn-north-1",
	}
}

// ParseCustomConfig 解析 京东云 DNS 自定义配置
func (p *JdcloudDNSConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// JdcloudDNSProvider 京东云 DNS 提供程序
type JdcloudDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg    *JdcloudDNSConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *JdcloudDNSCustomConfig
	Provider     struct {
		Auth   *core.Credential
		Client *client.DomainserviceClient
	}

	// 内部数据
	InternalData *provider.DNSProviderInternalData

	// 记录数据
	ItemData models.RRInfo
}

// NewJdcloudDNSProvider 创建京东云 DNS 提供程序
func NewJdcloudDNSProvider(ctx context.Context, logger *slog.Logger, cfg *JdcloudDNSConfig, opts ...Option) (provider.DNSProvider, error) {
	return NewJdcloudDNSProviderRaw(ctx, logger, cfg, opts...)
}

// NewJdcloudDNSProviderRaw 创建京东云 DNS 提供程序
func NewJdcloudDNSProviderRaw(ctx context.Context, logger *slog.Logger, cfg *JdcloudDNSConfig, opts ...Option) (*JdcloudDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &JdcloudDNSProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &JdcloudDNSCustomConfig{},
		InternalData: &provider.DNSProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 认证
	p.Provider.Auth = core.NewCredentials(
		p.cfg.AccessKey,
		p.cfg.SecretKey,
	)

	// 实例化
	p.Provider.Client = client.NewDomainserviceClient(p.Provider.Auth)

	// 如果提供了日志传输器，则设置HTTP客户端
	if p.opts.httpTransport != nil {
	}

	return p, nil
}

// Name 返回 DNS 提供程序的名称
func (p *JdcloudDNSProvider) Name() string {
	return "京东云 DNS"
}

// Check 检查子域名是否存在
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/describeresourcerecord
func (p *JdcloudDNSProvider) Check(ctx context.Context, domain string, subdomain string, newOriginValue string, recordType string) (data *provider.DNSProviderInternalData, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 查询域名记录
	requestDnsQuery := apis.NewDescribeResourceRecordRequestWithAllParams(
		p.cfg.Region,
		p.cfg.ZoneID,
		utils.IntToPtr(1),
		utils.IntToPtr(500),
		utils.StringToPtr(fullDomain),
	)
	responseDnsQuery, err := p.Provider.Client.DescribeResourceRecord(requestDnsQuery)
	if err != nil {
		return nil, fmt.Errorf("接口 DescribeResourceRecord 失败 %w", err)
	}
	// 响应参数
	if len(responseDnsQuery.Result.DataList) == 0 {
		p.InternalData.IsContinue = true
		return p.InternalData, nil
	}
	for _, item := range responseDnsQuery.Result.DataList {

		// 查找域名
		itemFullDomain := item.HostValue

		if utils.EqualDomainIgnoreCase(itemFullDomain, fullDomain) {

			// 记录数据
			p.ItemData = item

			// 查找记录类型
			itemRecordType := item.Type
			// 查找记录值
			itemOriginValue := item.HostValue

			// 检查记录类型
			if recordType != "" && !utils.EqualDomainIgnoreCase(itemRecordType, recordType) {
				continue
			}

			// 域名ID
			p.InternalData.RecordID = cast.ToString(item.Id)

			// 判断记录值
			if !utils.EqualDomainIgnoreCase(itemOriginValue, newOriginValue) {
				p.InternalData.IsContinue = true
				p.InternalData.Tips = fmt.Sprintf("记录值不匹配，期望 %v，实际 %v", newOriginValue, itemOriginValue)
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
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/createresourcerecord
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/modifyresourcerecord
func (p *JdcloudDNSProvider) Update(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if p.InternalData.RecordID == "" {

		// 创建域名记录
		requestDnsCreate := apis.NewCreateResourceRecordRequest(
			p.cfg.Region,
			p.cfg.ZoneID,
			&models.AddRR{
				HostRecord: fullDomain,  // 主机记录
				HostValue:  originValue, // 解析记录的值
				Type:       recordType,  // 解析的类型
			},
		)
		responseDnsCreate, err := p.Provider.Client.CreateResourceRecord(requestDnsCreate)
		if err != nil {
			return fmt.Errorf("接口 CreateResourceRecord 失败 %w", err)
		}

		// 域名ID
		p.InternalData.RecordID = cast.ToString(responseDnsCreate.Result.DataList.Id)

	} else if p.InternalData.RecordID != "" {

		// 更新域名记录
		requestDnsUpdate := apis.NewModifyResourceRecordRequest(
			p.cfg.Region,
			p.cfg.ZoneID,
			p.InternalData.RecordID,
			&models.UpdateRR{
				HostRecord: fullDomain,  // 主机记录
				HostValue:  originValue, // 解析记录的值
				Type:       recordType,  // 解析的类型
			},
		)
		_, err = p.Provider.Client.ModifyResourceRecord(requestDnsUpdate)
		if err != nil {
			return fmt.Errorf("接口 ModifyResourceRecord 失败 %w", err)
		}

	}

	return nil
}

// 京东云 DNS 自定义配置
type JdcloudDNSCustomConfig struct {
	IPGroupCode string `json:"ip_group_code,omitempty"` // IP 分组代码
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *JdcloudDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
