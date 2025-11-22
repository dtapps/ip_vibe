package config_helper

// Provider 定义单个提供商的配置结构
type Provider struct {
	Type        string            `mapstructure:"type"`                  // 类型：dns 或 cdn
	ID          string            `mapstructure:"id"`                    // 提供商的唯一ID
	AutoUpdate  bool              `mapstructure:"auto_update,omitempty"` // 是否自动更新
	Provider    string            `mapstructure:"provider"`              // 提供商名称，如 noip、cloudflare、edgeone、dnspod
	Domain      string            `mapstructure:"domain"`                // 域名
	IpType      string            `mapstructure:"ip_type"`               // IP类型，如 v4、v6
	Config      map[string]string `mapstructure:"config"`                // 提供商特定配置
	DNSProvider struct {
		Provider string            `mapstructure:"provider"` // 提供商名称，如 noip、cloudflare、edgeone、dnspod
		Config   map[string]string `mapstructure:"config"`   // 提供商特定配置
	} `mapstructure:"dns_provider,omitempty"` // CDN关联的DNS提供商ID
}

// Config 定义整个配置文件结构
type Config struct {
	Providers []Provider `mapstructure:"providers"`
}

// SafeGetConfig 安全获取提供商配置中的值
// 如果 provider 或 Config 为 nil，或 key 不存在，或值为空，均返回空字符串
func (p *Provider) SafeGetConfig(key string) string {
	if p == nil || p.Config == nil {
		return ""
	}
	v, ok := p.Config[key]
	if !ok || v == "" {
		return ""
	}
	return v
}

// SafeGetConfigWithDefault 安全获取提供商配置中的值，如果不存在或为空，则返回默认值
func (p *Provider) SafeGetConfigWithDefault(key string, defaultValue string) string {
	if p == nil || p.Config == nil {
		return defaultValue
	}
	v, ok := p.Config[key]
	if !ok || v == "" {
		return defaultValue
	}
	return v
}
