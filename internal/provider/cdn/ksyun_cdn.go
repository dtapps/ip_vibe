package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/provider"

	cdn "github.com/kingsoftcloud/sdk-go/v2/ksyun/client/cdn/v3"
	"github.com/kingsoftcloud/sdk-go/v2/ksyun/common"
	"github.com/kingsoftcloud/sdk-go/v2/ksyun/common/profile"
)

// KsyunCDNConfig 金山云 CDN 配置
type KsyunCDNConfig struct {
	Region     string `json:"region,omitempty"` // 区域
	AccessKey  string `json:"access_key"`       // 密钥ID
	SecretKey  string `json:"secret_key"`       // 密钥Key
	DomainName string `json:"domain_name"`      // 主域名
}

// Validate 验证金山云 CDN 配置
func (cfg *KsyunCDNConfig) Validate() error {
	if cfg.Region == "" {
		return fmt.Errorf("Region 不能为空")
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

// NewDefaultKsyunCDNConfig 创建默认的金山云 CDN 配置
func NewDefaultKsyunCDNConfig() *KsyunCDNConfig {
	return &KsyunCDNConfig{
		Region: "cn-north-1",
	}
}

// ParseCustomConfig 解析 金山云 CDN 自定义配置
func (p *KsyunCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// KsyunCDNProvider 金山云 CDN 提供程序
type KsyunCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *KsyunCDNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *KsyunCDNCustomConfig
	Provider     struct {
		Auth   *common.Credential
		Cfg    *profile.ClientProfile
		Client *cdn.Client
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	// ItemData teo.AccelerationDomain
}

// NewKsyunCDNProvider 创建金山云 CDN 提供程序
func NewKsyunCDNProvider(ctx context.Context, logger *slog.Logger, cfg *KsyunCDNConfig, opts ...Option) (provider.CDNProvider, error) {
	return NewKsyunCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewKsyunCDNProviderRaw 创建金山云 CDN 提供程序
func NewKsyunCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *KsyunCDNConfig, opts ...Option) (*KsyunCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	var err error
	p := &KsyunCDNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &KsyunCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth = common.NewCredential(
		p.cfg.AccessKey,
		p.cfg.SecretKey,
	)

	// 客户端配置
	p.Provider.Cfg = profile.NewClientProfile()
	p.Provider.Cfg.HttpProfile.ReqMethod = "POST|GET"
	p.Provider.Cfg.HttpProfile.ReqTimeout = 60
	p.Provider.Cfg.HttpProfile.Endpoint = "cdn.api.ksyun.com"

	// 实例化
	p.Provider.Client, err = cdn.NewClient(
		p.Provider.Auth,
		p.cfg.Region,
		p.Provider.Cfg,
	)
	if err != nil {
		return nil, fmt.Errorf("实例化失败 %w", err)
	}

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *KsyunCDNProvider) Name() string {
	return "金山云 CDN"
}

// 金山云 CDN 自定义配置
type KsyunCDNCustomConfig struct {
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *KsyunCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
