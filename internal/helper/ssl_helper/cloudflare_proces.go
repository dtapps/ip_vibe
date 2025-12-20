package ssl_helper

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/casbin/ent-adapter/ent"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
	"go.dtapp.net/library/contrib/http_log"
)

// 实现 SSLProcess 接口
type SSLProcessCloudflareImpl struct {
	debug         bool                       // 调试模式
	logger        *slog.Logger               // 日志
	db            *ent.Client                // 数据库
	httpTransport *http_log.LoggingTransport // HTTP日志记录器

	platformConfig CloudflareConfigure // 平台配置

	legoConfig *LegoConfig  // Lego配置
	client     *lego.Client // ACME 客户端
}

// NewSSLProcessCloudflare 创建实例
func NewSSLProcessCloudflare(debug bool, logger *slog.Logger, db *ent.Client, httpTransport *http_log.LoggingTransport) SSLProcess {
	return &SSLProcessCloudflareImpl{
		debug:         debug,
		logger:        logger,
		db:            db,
		httpTransport: httpTransport,
	}
}

type CloudflareConfigure struct {
	ApiKey string `json:"api_key"` // API密钥
	Email  string `json:"email"`   // 邮箱
}

// 配置
func (d *SSLProcessCloudflareImpl) Configure(ctx context.Context, params *SSLProcessParams) error {

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
	config.HTTPClient.Transport = d.httpTransport

	// 创建 ACME 客户端
	d.client, err = lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("无法创建 ACME 客户端: %w", err)
	}

	return nil
}

// 申请
func (d *SSLProcessCloudflareImpl) Apply(ctx context.Context, params *SSLProcessParams) error {

	// 要申请的域名列表
	domains := []string{"example.com", "*.example.com"} // 可以同时申请主域名和通配符域名

	// 设置 DNS-01 挑战提供者 (Cloudflare)
	cfProvider, err := cloudflare.NewDNSProviderConfig(&cloudflare.Config{
		AuthToken: d.platformConfig.ApiKey,
	})
	if err != nil {
		return fmt.Errorf("无法创建 Cloudflare DNS 提供者: %w", err)
	}

	// 将 DNS Provider 传递给 lego 客户端
	err = d.client.Challenge.SetDNS01Provider(cfProvider)
	if err != nil {
		return fmt.Errorf("无法设置 DNS-01 提供者: %w", err)
	}

	// 向 CA 注册账户
	// TermsOfServiceAgreed: true 表示自动同意服务条款
	reg, err := d.client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return fmt.Errorf("无法注册 ACME 账户: %w", err)
	}
	d.legoConfig.Registration = reg
	fmt.Println("ACME 账户注册成功！")

	// 请求证书
	fmt.Printf("正在为域名 %v 申请证书...\n", domains)
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true, // 是否将中间证书一起返回
	}

	certificates, err := d.client.Certificate.Obtain(request)
	if err != nil {
		return fmt.Errorf("无法获取证书: %w", err)
	}

	// 7. 保存证书和私钥到文件
	err = os.WriteFile("certificates/"+domains[0]+"_fullchain.pem", certificates.Certificate, 0644)
	if err != nil {
		return fmt.Errorf("无法写入证书文件: %w", err)
	}
	err = os.WriteFile("certificates/"+domains[0]+"_private.key", certificates.PrivateKey, 0600) // 私钥权限设为 600
	if err != nil {
		return fmt.Errorf("无法写入私钥文件: %w", err)
	}

	fmt.Println("🎉 证书申请成功！")
	fmt.Printf("证书已保存到: certificates/%s_fullchain.pem\n", domains[0])
	fmt.Printf("私钥已保存到: certificates/%s_private.key\n", domains[0])

	return nil
}
