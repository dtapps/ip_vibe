package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	if err := json.Unmarshal(config_json, p); err != nil {
		return err
	}
	return nil
}

// JdcloudDNSProvider 京东云 DNS 提供程序
type JdcloudDNSProvider struct {
	provider.UnimplementedDNSProvider // 嵌入未实现的

	cfg  *JdcloudDNSConfig
	opts *dnsOptions

	CustomConfig *JdcloudDNSCustomConfig
	Provider     struct {
		Auth   *core.Credential
		Client *client.DomainserviceClient
	}
}

// NewJdcloudDNSProvider 创建京东云 DNS 提供程序
func NewJdcloudDNSProvider(ctx context.Context, cfg *JdcloudDNSConfig, opts ...DnsOption) (provider.DNSProvider, error) {
	return NewJdcloudDNSProviderRaw(ctx, cfg, opts...)
}

// NewJdcloudDNSProviderRaw 创建京东云 DNS 提供程序
func NewJdcloudDNSProviderRaw(ctx context.Context, cfg *JdcloudDNSConfig, opts ...DnsOption) (*JdcloudDNSProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &JdcloudDNSProvider{
		cfg:          cfg,
		opts:         &dnsOptions{},
		CustomConfig: &JdcloudDNSCustomConfig{},
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

// Check 检查子域名是否存在
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/describeresourcerecord
func (p *JdcloudDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	// 请求参数
	request := apis.NewDescribeResourceRecordRequestWithAllParams(
		p.cfg.Region,
		p.cfg.ZoneID,
		utils.IntToPtr(1),
		utils.IntToPtr(500),
		utils.StringToPtr(fullDomain),
	)
	// 请求
	response, err := p.Provider.Client.DescribeResourceRecord(request)
	printJSON("查询域名记录", request, response)
	if err != nil {
		return
	}
	// 响应参数
	for _, item := range response.Result.DataList {
		if strings.EqualFold(utils.TrimDots(item.HostValue), utils.TrimDots(fullDomain)) {
			if recordType != "" && !strings.EqualFold(item.Type, recordType) {
				continue
			}
			recordID = cast.ToString(item.Id)
			value = item.HostValue
			return recordID, value, nil
		}
	}

	return "", "", provider.ErrRecordNotFound
}

// Update 更新 DNS 记录
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/createresourcerecord
// https://docs.jdcloud.com/cn/jd-cloud-dns/api/modifyresourcerecord
func (p *JdcloudDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {

	// 完整域名
	fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)

	if recordID == "" {
		// 请求参数
		request := apis.NewCreateResourceRecordRequest(
			p.cfg.Region,
			p.cfg.ZoneID,
			&models.AddRR{
				HostRecord: fullDomain, // 主机记录
				HostValue:  value,      // 解析记录的值
				Type:       recordType, // 解析的类型
			},
		)
		// 请求
		response, err := p.Provider.Client.CreateResourceRecord(request)
		printJSON("创建域名记录", request, response)
		if err != nil {
			return err
		}
	}

	if recordID != "" {
		// 请求参数
		request := apis.NewModifyResourceRecordRequest(
			p.cfg.Region,
			p.cfg.ZoneID,
			recordID,
			&models.UpdateRR{
				HostRecord: fullDomain, // 主机记录
				HostValue:  value,      // 解析记录的值
				Type:       recordType, // 解析的类型
			},
		)
		// 请求
		response, err := p.Provider.Client.ModifyResourceRecord(request)
		printJSON("更新域名记录", request, response)
		if err != nil {
			return err
		}
	}

	return nil
}

// Save 保存 DNS 记录
func (p *JdcloudDNSProvider) Save(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error) {
	return nil
}

// Name 返回 DNS 提供程序的名称
func (p *JdcloudDNSProvider) Name() string {
	return "京东云 DNS"
}

// 京东云 DNS 自定义配置
type JdcloudDNSCustomConfig struct {
}

// ParseCustomConfig 解析 DNS 自定义配置
func (p *JdcloudDNSProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	return nil
}
