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
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/carlmjohnson/requests"
	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
	"github.com/levigross/grequests/v2"
	fakeUserAgent "github.com/lib4u/fake-useragent"
	"go.dtapp.net/library/contrib/http_log"
	"resty.dev/v3"
)

type DomainCheck struct {
	useragent     string                     // 浏览器标识
	logger        *slog.Logger               // 日志记录器
	httpTransport *http_log.LoggingTransport // HTTP日志记录器
	debug         bool                       // 是否开启调试模式
}

func NewDomainCheck(logger *slog.Logger, httpTransport *http_log.LoggingTransport, debug bool) *DomainCheck {
	var useragent string
	ua, err := fakeUserAgent.New()
	if err == nil {
		useragent = ua.GetRandom()
	}
	return &DomainCheck{
		useragent:     useragent,
		logger:        logger,
		httpTransport: httpTransport,
		debug:         debug,
	}
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

// Check 执行实际的网络检查并返回结果
func (s *DomainCheck) Check(ctx context.Context, domain string, protocol string) *CheckResult {
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

	// ---------- DNS 解析 ----------
	if ips, err := net.LookupIP(u.Host); err == nil {
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
		if err := s.fetchSSLInfo(ctx, u.Host, result); err != nil {
			result.StatusMsg = "TLS 握手失败"
			result.ErrorLog = err.Error()
			result.Duration = time.Since(start)
			return result
		}
	}

	// ---------- HTTP / HTTPS 请求 ----------
	fetchMethods := []fetchFunc{
		s.fetchHTTP,              // 官方 标准 库
		s.fetchHTTPWithResty,     // 第三方 Resty 库
		s.fetchHTTPWithGrequests, // 第三方 Grequests 库
		s.fetchHTTPWithReq,       // 第三方 Req 库
		s.fetchHTTPWithRequests,  // 第三方 Requests 库
		s.fetchHTTPWithColly,     // 第三方 Colly 库
	}

	// 随机选一个
	selectedIndex := rand.Intn(len(fetchMethods))
	selectedMethod := fetchMethods[selectedIndex]
	if err := selectedMethod(ctx, u, start, result); err != nil {
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

type fetchFunc func(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error

// 发起 HTTP 请求，使用官方 标准 库
func (s *DomainCheck) fetchHTTP(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeStandard

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		client.Transport = s.httpTransport.Instance(nil)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		result.StatusMsg = "请求创建失败"
		result.ErrorLog = err.Error()
		result.Duration = time.Since(start)
		return err
	}

	// 添加浏览器标识
	// req.Header.Set("User-Agent", s.useragent)
	// s.logger.Debug("请求浏览器标识",
	// 	slog.String("useragent", s.useragent),
	// )

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
func (s *DomainCheck) fetchHTTPWithResty(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeResty

	// 创建 Resty 客户端
	client := resty.New()
	defer client.Close()

	// 设置超时
	client.SetTimeout(60 * time.Second)

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		client.SetTransport(s.httpTransport.Instance(nil))
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

// 发起 HTTP 请求，使用第三方 Grequests 库
func (s *DomainCheck) fetchHTTPWithGrequests(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeGrequests

	// 组装 RequestOptions
	opts := &grequests.RequestOptions{
		Context: ctx,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		opts.HTTPClient.Transport = s.httpTransport.Instance(nil)
	}

	// 发起 GET 请求
	resp, err := grequests.Get(ctx, u.String(), grequests.FromRequestOptions(opts))
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP(Grequests) 请求失败 " + u.String()
		result.ErrorLog = err.Error()
		return err
	}
	defer resp.Close()

	result.HTTPStatus = resp.StatusCode

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsSuccess = true
		result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode)
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	return nil
}

// 发起 HTTP 请求，使用第三方 Req 库
// https://req.cool/
func (s *DomainCheck) fetchHTTPWithReq(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeReq

	// 创建 Req 客户端
	client := req.NewClient()

	// 设置超时
	client.SetTimeout(60 * time.Second)

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		transport := client.GetTransport()
		transport.WrapRoundTrip(func(rt http.RoundTripper) http.RoundTripper {
			return s.httpTransport.Middleware()(rt)
		})
	}

	// 伪装 HTTP 指纹
	// client.ImpersonateChrome()

	// 发起 GET 请求
	resp, err := client.R().SetContext(ctx).Get(u.String())
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

// 发起 HTTP 请求，使用第三方 Requests 库
func (s *DomainCheck) fetchHTTPWithRequests(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeReq

	// 创建 Requests 客户端
	client := requests.URL(u.String()).
		Method(http.MethodGet)

	// 设置超时
	client.Client(&http.Client{
		Timeout: 60 * time.Second,
	})

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		client.Transport(s.httpTransport.Instance(nil))
	}

	// 发起 GET 请求
	err := client.Fetch(ctx)
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP(Req) 请求失败 " + u.String()
		result.ErrorLog = err.Error()
		return err
	}

	// result.HTTPStatus = resp.StatusCode

	// if resp.StatusCode >= 200 && resp.StatusCode < 400 {
	// 	result.IsSuccess = true
	// 	result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode)
	// } else {
	// 	result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode)
	// }

	return nil
}

// 发起 HTTP 请求，使用第三方 Colly 库
// https://go-colly.org/
func (s *DomainCheck) fetchHTTPWithColly(ctx context.Context, u *url.URL, start time.Time, result *CheckResult) error {

	// 设置检查类型
	result.CheckType = schematype.DomainMonitorHistoryCheckTypeColly

	// 创建 Colly 采集器
	client := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true), // 避免阻塞
	)

	// 设置超时
	client.SetRequestTimeout(60 * time.Second)

	// 添加 HTTP 日志记录器
	if s.httpTransport != nil {
		client.WithTransport(s.httpTransport.Instance(nil))
	}

	// 初始化结果变量（通过闭包捕获）
	var statusCode int

	// 注册回调
	client.OnResponse(func(r *colly.Response) {
		statusCode = r.StatusCode
	})

	// 启动请求
	err := client.Request(http.MethodGet, u.String(), nil, nil, nil)
	result.Duration = time.Since(start)

	if err != nil {
		result.StatusMsg = "HTTP(Colly) 请求失败 " + u.String()
		result.ErrorLog = err.Error()
		return err
	}

	// 等待异步请求完成（Colly 默认异步）
	client.Wait()

	result.HTTPStatus = statusCode

	if statusCode >= 200 && statusCode < 400 {
		result.IsSuccess = true
		result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", statusCode)
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", statusCode)
	}

	return nil
}
