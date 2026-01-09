package domain_helper

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	fakeUserAgent "github.com/lib4u/fake-useragent"
)

type DomainCheck struct {
	useragent string       // 浏览器标识
	logger    *slog.Logger // 日志记录器
}

func NewDomainCheck(logger *slog.Logger) *DomainCheck {
	var useragent string
	ua, err := fakeUserAgent.New()
	if err == nil {
		useragent = ua.GetRandom()
	}
	return &DomainCheck{
		useragent: useragent,
		logger:    logger,
	}
}

// CheckResult 包含本次检查的所有结果
type CheckResult struct {
	IsSuccess  bool          // 是否成功
	Duration   time.Duration // 检查耗时
	StatusMsg  string        // 状态信息
	HTTPStatus int           // HTTP状态码
	ErrorLog   string        // 错误日志

	IPResolved []string // 解析的IP地址

	// SSL 证书信息 (HTTP 检查时为空)
	SSLNotDate          time.Time // 生效日期
	SSLExpiryDate       time.Time // 过期日期
	SSLCommonName       string    // 通用名
	SSLIssuerName       string    // 颁发者
	SSLSignatureAlg     string    // 签名算法
	SSLSANList          string    // 主体别名列表
	SSLCertificateChain string    // 证书链内容 (PEM格式)
	SSLSerialNumber     string    // 序列号
	SSLVersion          int       // 版本
}

// Check 执行实际的网络检查并返回结果
func (s *DomainCheck) Check(ctx context.Context, domain string, protocol string) *CheckResult {
	start := time.Now()

	result := &CheckResult{
		StatusMsg: "未执行",
	}

	// ---------- DNS 解析 ----------
	if ips, err := net.LookupIP(domain); err == nil {
		for _, ip := range ips {
			result.IPResolved = append(result.IPResolved, ip.String())
		}
	} else {
		result.StatusMsg = "DNS解析失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// ---------- HTTPS：单独获取 SSL 信息 ----------
	if protocol == "https" {
		if err := s.fetchSSLInfo(ctx, domain, result); err != nil {
			result.StatusMsg = "TLS 握手失败"
			result.ErrorLog = err.Error()
			result.Duration = time.Since(start)
			return result
		}
	}

	// ---------- HTTP / HTTPS 请求 ----------
	if err := s.fetchHTTP(ctx, domain, protocol, start, result); err != nil {
		return result
	}

	return result
}

// 获取 SSL 信息
func (s *DomainCheck) fetchSSLInfo(ctx context.Context, domain string, result *CheckResult) error {
	dialer := &net.Dialer{Timeout: 60 * time.Second}

	conn, err := tls.DialWithDialer(dialer, "tcp", domain+":443", &tls.Config{
		ServerName:         domain,
		InsecureSkipVerify: true, // 仅用于采集证书
	})
	if err != nil {
		return err
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("未获取到证书")
	}

	cert := state.PeerCertificates[0]
	result.SSLNotDate = cert.NotBefore
	result.SSLExpiryDate = cert.NotAfter
	result.SSLCommonName = cert.Subject.CommonName
	result.SSLIssuerName = cert.Issuer.CommonName
	result.SSLSignatureAlg = cert.SignatureAlgorithm.String()
	result.SSLSANList = strings.Join(cert.DNSNames, ",")

	var chain strings.Builder
	for _, c := range state.PeerCertificates {
		chain.Write(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: c.Raw,
		}))
	}
	result.SSLCertificateChain = chain.String()

	result.SSLSerialNumber = cert.SerialNumber.String()
	result.SSLVersion = cert.Version

	return nil
}

// 发起 HTTP 请求，使用官方 http.Client
func (s *DomainCheck) fetchHTTP(ctx context.Context, protocol string, domain string, start time.Time, result *CheckResult) error {

	url := protocol + "://" + domain

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.StatusMsg = "请求创建失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return err
	}

	// 添加浏览器标识
	req.Header.Set("User-Agent", s.useragent)
	s.logger.Debug("请求浏览器标识",
		slog.String("useragent", s.useragent),
	)

	resp, err := client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP 请求失败"
		result.ErrorLog = err.Error()
		return err
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsSuccess = true
		result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode)
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	return nil
}
