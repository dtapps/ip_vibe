package ssl_helper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/westcn"
	"github.com/go-acme/lego/v4/registration"
	"go.dtapp.net/library/contrib/http_log"
)

// 客户端配置
type WestcnConfigure struct {
	Username string `json:"username"` // 用户名
	Password string `json:"password"` // 密码
	Email    string `json:"email"`    // 邮箱
}

// 实现 SSLProcess 接口
type SSLProcessWestcnImpl struct {
	logger        *slog.Logger               // 日志
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 调试模式

	platformConfig WestcnConfigure // 平台配置

	legoConfig *LegoConfig  // Lego配置
	client     *lego.Client // ACME 客户端
}

// NewSSLProcessWestcn 创建实例
func NewSSLProcessWestcn(logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) SSLProcess {
	return &SSLProcessWestcnImpl{
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
}

// 配置
func (d *SSLProcessWestcnImpl) Configure(ctx context.Context, params *SSLProcessParams) error {

	err := json.Unmarshal(params.ConfigJSON, &d.platformConfig)
	if err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	d.logger.InfoContext(ctx, "配置",
		slog.String("邮箱", d.platformConfig.Email),
	)

	// 生成私钥
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("无法生成私钥: %w", err)
	}

	// 准备 Lego 配置
	d.legoConfig = &LegoConfig{
		Email: d.platformConfig.Email, // 用于 Let's Encrypt 账户
		key:   privateKey,
	}

	// 配置 ACME 客户端
	config := lego.NewConfig(d.legoConfig)
	// 使用 Let's Encrypt 生产环境
	config.CADirURL = lego.LEDirectoryProduction
	// 使用 Let's Encrypt 测试环境 (推荐初次运行时使用，无请求频率限制)
	// config.CADirURL = lego.LEDirectoryStaging
	config.Certificate.KeyType = certcrypto.RSA2048

	// 设置 HTTP 客户端传输 (用于记录日志)
	if d.httpTransport != nil {
		config.HTTPClient.Transport = d.httpTransport
	}

	// 创建 ACME 客户端
	d.client, err = lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("无法创建 ACME 客户端: %w", err)
	}

	return nil
}

// 申请
// domains 域名列表，可以同时申请主域名和通配符域名
func (d *SSLProcessWestcnImpl) Apply(ctx context.Context, domains []string) (res *certificate.Resource, err error) {

	d.logger.InfoContext(ctx, "设置 DNS-01 提供者")

	// 设置 DNS-01 提供者
	dnsProviderConfig := westcn.NewDefaultConfig()
	dnsProviderConfig.Username = d.platformConfig.Username
	dnsProviderConfig.Password = d.platformConfig.Password
	cfProvider, err := westcn.NewDNSProviderConfig(dnsProviderConfig)
	if err != nil {
		return nil, fmt.Errorf("无法创建 DNS-01 提供者: %w", err)
	}

	// 将 DNS Provider 传递给 lego 客户端
	err = d.client.Challenge.SetDNS01Provider(cfProvider)
	if err != nil {
		return nil, fmt.Errorf("无法设置 DNS-01 提供者: %w", err)
	}

	d.logger.InfoContext(ctx, "向 CA 注册账户")

	// 向 CA 注册账户
	// TermsOfServiceAgreed: true 表示自动同意服务条款
	reg, err := d.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("无法注册 ACME 账户: %w", err)
	}
	d.legoConfig.Registration = reg

	d.logger.InfoContext(ctx, "获取 SSL 证书")

	// 获取 SSL 证书
	res, err = d.client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true, // 是否将中间证书一起返回
	})
	if err != nil {
		return nil, fmt.Errorf("无法获取证书: %w", err)
	}

	return res, nil
}
