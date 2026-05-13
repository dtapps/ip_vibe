package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"vibe_ip/internal/provider"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/cdn/client"
)

// JdcloudCDNConfig 京东云 CDN 配置
type JdcloudCDNConfig struct {
	Region     string `json:"region,omitempty"` // 区域
	AccessKey  string `json:"access_key"`       // 密钥ID
	SecretKey  string `json:"secret_key"`       // 密钥密钥
	DomainName string `json:"domain_name"`      // 主域名
}

// DefaultValidate 验证配置
func (cfg *JdcloudCDNConfig) DefaultValidate() error {
	if cfg.Region == "" {
		return fmt.Errorf("region 不能为空")
	}
	if cfg.AccessKey == "" {
		return fmt.Errorf("AccessKey 不能为空")
	}
	if cfg.SecretKey == "" {
		return fmt.Errorf("SecretKey 不能为空")
	}
	return nil
}

// CdnValidate 验证配置
func (cfg *JdcloudCDNConfig) CdnValidate() error {
	if cfg.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	return nil
}

// NewDefaultJdcloudCDNConfig 创建默认的京东云 CDN 配置
func NewDefaultJdcloudCDNConfig() *JdcloudCDNConfig {
	return &JdcloudCDNConfig{
		Region: "cn-north-1",
	}
}

// ParseCustomConfig 解析 京东云 CDN 自定义配置
func (p *JdcloudCDNConfig) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}

// JdcloudCDNProvider 京东云 CDN 提供程序
type JdcloudCDNProvider struct {
	provider.UnimplementedCDNProvider // 嵌入未实现的

	cfg    *JdcloudCDNConfig
	opts   *Options
	logger *slog.Logger

	CustomConfig *JdcloudCDNCustomConfig
	Provider     struct {
		Auth   *core.Credential
		Client *client.CdnClient
	}

	// 内部数据
	InternalData *provider.CDNProviderInternalData

	// 记录数据
	// ItemData teo.AccelerationDomain
}

// NewJdcloudCDNProvider 创建京东云 CDN 提供程序
func NewJdcloudCDNProvider(ctx context.Context, logger *slog.Logger, cfg *JdcloudCDNConfig, opts ...Option) (provider.CDNProvider, error) {
	return NewJdcloudCDNProviderRaw(ctx, logger, cfg, opts...)
}

// NewJdcloudCDNProviderRaw 创建京东云 CDN 提供程序
func NewJdcloudCDNProviderRaw(ctx context.Context, logger *slog.Logger, cfg *JdcloudCDNConfig, opts ...Option) (*JdcloudCDNProvider, error) {

	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := cfg.DefaultValidate(); err != nil {
		return nil, fmt.Errorf("配置无效: %w", err)
	}

	p := &JdcloudCDNProvider{
		cfg:          cfg,
		opts:         &Options{},
		logger:       logger,
		CustomConfig: &JdcloudCDNCustomConfig{},
		InternalData: &provider.CDNProviderInternalData{},
	}

	// 应用选项
	for _, opt := range opts {
		opt(p.opts)
	}

	// 密钥信息
	p.Provider.Auth = core.NewCredentials(
		p.cfg.AccessKey,
		p.cfg.SecretKey,
	)

	// 实例化
	p.Provider.Client = client.NewCdnClient(p.Provider.Auth)

	return p, nil
}

// Name 返回 CDN 提供程序的名称
func (p *JdcloudCDNProvider) Name() string {
	return "京东云 CDN"
}

// 京东云 CDN 自定义配置
type JdcloudCDNCustomConfig struct {
}

// ParseCustomConfig 解析 CDN 自定义配置
func (p *JdcloudCDNProvider) ParseCustomConfig(config_json json.RawMessage) (err error) {
	if len(config_json) == 0 {
		return nil
	}
	if err := json.Unmarshal(config_json, p.CustomConfig); err != nil {
		return fmt.Errorf("解析自定义配置失败 %w", err)
	}
	return nil
}
