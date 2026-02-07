package domain_helper

import (
	"context"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"vibe_ip/internal/core"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/helper/sys_env"

	"github.com/imroc/req/v3"
	"go.dtapp.net/library/contrib/http_log"
	"resty.dev/v3"
)

type DomainCheck struct {
	logger        *slog.Logger                  // 日志记录器
	httpTransport *http_log.LoggingRoundTripper // HTTP日志记录器
	debug         bool                          // 是否开启调试模式
	timeout       time.Duration                 // 请求超时时间
	httpClient    *http.Client                  // HTTP 客户端
	restyClient   *resty.Client                 // Resty 客户端
	reqClient     *req.Client                   // Req 客户端
}

// NewDomainCheck 创建一个新的 DomainCheck 实例
func NewDomainCheck(logger *slog.Logger, httpTransport *http_log.LoggingRoundTripper) *DomainCheck {
	debug := sys_env.GvaEnv.DEBUG
	s := &DomainCheck{
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
		timeout:       60 * time.Second,
	}
	s.initHttpClient()
	return s
}

// initHttpClient 初始化 HTTP 客户端
func (s *DomainCheck) initHttpClient() {

	// 创建客户端
	httpClient := &http.Client{}
	restyClient := resty.New()
	reqClient := req.NewClient()

	// 开启日志
	if s.debug {
		restyClient.EnableDebug()
		reqClient.DevMode()
	}

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		// 	httpClient.Transport = s.httpTransport.Instance(nil)
		// 	restyClient.SetTransport(s.httpTransport.Instance(restyClient.Transport()))
		// 	reqClient.GetTransport().WrapRoundTrip(func(rt http.RoundTripper) http.RoundTripper {
		// 		return s.httpTransport.Instance(rt)
		// 	})
		// }
		// if !s.debug && s.httpTransport != nil {
		httpClient.Transport = s.httpTransport.InstanceNoBody(nil)
		restyClient.SetTransport(s.httpTransport.InstanceNoBody(restyClient.Transport()))
		reqClient.GetTransport().WrapRoundTrip(func(rt http.RoundTripper) http.RoundTripper {
			return s.httpTransport.InstanceNoBody(rt)
		})
	}

	// 设置超时
	httpClient.Timeout = s.timeout
	restyClient.SetTimeout(s.timeout)
	reqClient.SetTimeout(s.timeout)

	// 设置 HTTP 客户端
	s.httpClient = httpClient
	s.restyClient = restyClient
	s.reqClient = reqClient
}

// CheckResult 包含本次检查的所有结果
type CheckResult struct {
	IsSuccess  bool          // 是否成功
	Duration   time.Duration // 检查耗时
	StatusMsg  string        // 状态信息
	HTTPStatus int           // HTTP状态码
	ErrorLog   string        // 错误日志

	CheckType string // 检查类型

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

type fetchHttpFunc func(ctx context.Context, u *url.URL, httpHeaders map[string]string, start time.Time, result *CheckResult) error

// Check 执行实际的网络检查并返回结果
// protocol: 协议，例如 "http" 或 "https"
// domain: 要检查的域名
// headers: HTTP 请求头，格式为 "Key=Value"
func (s *DomainCheck) Check(ctx context.Context, protocol string, domain string, headers []string) *CheckResult {
	start := time.Now()

	result := &CheckResult{
		StatusMsg: "未执行",
		CheckType: schematype.DomainMonitorHistoryCheckTypeUnknown,
	}

	// 把 protocol + domain 解析成 URL
	u, err := url.Parse(fmt.Sprintf("%s://%s", protocol, domain))
	if err != nil {
		result.StatusMsg = "URL 解析失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	httpHeaders := make(map[string]string)
	for _, h := range headers {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) == 2 {
			httpHeaders[parts[0]] = parts[1]
		}
	}

	// ---------- DNS 解析 ----------
	if ips, err := core.GetDNSResultList(ctx, domain, net.JoinHostPort("119.29.29.29", "53")); err == nil {
		for _, ip := range ips {
			result.IPResolved = append(result.IPResolved, ip)
		}
	} else {
		result.StatusMsg = "DNS解析失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	// ---------- HTTPS：单独获取 SSL 信息 ----------
	if protocol == "https" {
		if err := s.fetchSSLInfo(ctx, u.Host, result); err != nil {
			result.StatusMsg = "TLS 握手失败"
			result.ErrorLog = err.Error()
			result.Duration = time.Since(start)
			return result
		}
	}

	// ---------- HTTP / HTTPS 请求 ----------
	fetchHttps := []fetchHttpFunc{
		s.fetchHTTPWithStandard, // 官方 标准 库
		s.fetchHTTPWithResty,    // 第三方 Resty 库
		s.fetchHTTPWithReq,      // 第三方 Req 库
	}
	// 随机选一个
	fetchHttp := fetchHttps[rand.Intn(len(fetchHttps))]
	if err := fetchHttp(ctx, u, httpHeaders, start, result); err != nil {
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

// 发起 HTTP 请求，使用官方 标准 库
func (s *DomainCheck) fetchHTTPWithStandard(ctx context.Context, u *url.URL, httpHeaders map[string]string, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeStandard

	// 复制 HTTP 客户端
	client := s.httpClient

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		result.StatusMsg = "请求创建失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return err
	}

	// 添加自定义 HTTP 请求头
	for k, v := range httpHeaders {
		req.Header.Set(k, v)
	}

	// 发起 GET 请求
	resp, err := client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP 请求失败 " + u.String()
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

// 发起 HTTP 请求，使用第三方 Resty 库
// https://resty.dev/
func (s *DomainCheck) fetchHTTPWithResty(ctx context.Context, u *url.URL, httpHeaders map[string]string, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeResty

	// 复制 Resty 客户端
	client := s.restyClient.Clone(ctx)

	// 添加自定义 HTTP 请求头
	for k, v := range httpHeaders {
		client.SetHeader(k, v)
	}

	// 发起 GET 请求
	resp, err := client.R().SetContext(ctx).Get(u.String())
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP(Resty) 请求失败 " + u.String()
		result.ErrorLog = err.Error()
		return err
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode()

	if resp.StatusCode() >= 200 && resp.StatusCode() < 400 {
		result.IsSuccess = true
		result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode())
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode())
	}

	return nil
}

// type fetchImpersonateFunc func() *req.Client

// 发起 HTTP 请求，使用第三方 Req 库
// https://req.cool/
func (s *DomainCheck) fetchHTTPWithReq(ctx context.Context, u *url.URL, httpHeaders map[string]string, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeReq

	// 复制 Req 客户端
	client := s.reqClient.Clone()

	// 伪装 HTTP 指纹
	// fetchImpersonates := []fetchImpersonateFunc{
	// 	client.ImpersonateChrome,
	// 	client.ImpersonateFirefox,
	// 	client.ImpersonateSafari,
	// }

	// 随机选一个
	// selectedImpersonateIndex := rand.Intn(len(fetchImpersonates))
	// selectedImpersonateMethod := fetchImpersonates[selectedImpersonateIndex]
	// selectedImpersonateMethod()

	// 创建 Req 请求
	req := client.R().SetContext(ctx)

	// 添加自定义 HTTP 请求头
	for k, v := range httpHeaders {
		req.SetHeader(k, v)
	}

	// 发起 GET 请求
	resp, err := req.Get(u.String())
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP(Req) 请求失败 " + u.String()
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
