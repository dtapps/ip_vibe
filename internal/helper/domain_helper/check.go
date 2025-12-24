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
)

type DomainCheck struct {
	logger *slog.Logger // 日志记录器
}

func NewDomainCheck(logger *slog.Logger) *DomainCheck {
	return &DomainCheck{
		logger: logger,
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
	SSLExpiryDate    time.Time // 证书过期日期
	SSLCommonName    string    // 证书通用名
	SSLIssuerName    string    // 证书颁发者
	SSLSignatureAlg  string    // 签名算法
	SANList          string    // SAN别名列表
	CertificateChain string    // 证书链内容 (PEM格式)
}

// Check 执行实际的网络检查并返回结果
func (s *DomainCheck) Check(ctx context.Context, domain string, protocol string) *CheckResult {
	result := &CheckResult{
		StatusMsg: "未执行",
	}

	url := protocol + "://" + domain
	start := time.Now()

	// 解析域名获取IP地址
	ips, err := net.LookupIP(domain)
	if err != nil {
		result.ErrorLog = fmt.Sprintf("DNS解析失败: %v", err)
		result.StatusMsg = "DNS解析失败"
		result.Duration = time.Since(start)
	} else {
		ipStrs := make([]string, len(ips))
		for i, ip := range ips {
			ipStrs[i] = ip.String()
		}
		result.IPResolved = ipStrs
	}

	// 确保域名包含端口
	addr := domain
	if !strings.Contains(addr, ":") {
		if protocol == "https" {
			addr += ":443"
		} else {
			addr += ":80"
		}
	}

	s.logger.InfoContext(ctx, "确保域名包含端口",
		slog.String("地址", addr),
	)

	// --- 统一的 Transport 来处理 HTTPS 证书获取和 HTTP 请求 ---
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	client := http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// 如果是 HTTPS，我们需要包装连接来获取证书
				if protocol == "https" {
					// 建立 TCP 连接
					tcpConn, err := dialer.DialContext(ctx, "tcp", addr)
					if err != nil {
						return nil, err
					}

					// 升级为 TLS 连接
					tlsConn := tls.Client(tcpConn, &tls.Config{
						ServerName:         domain,
						InsecureSkipVerify: true, // 忽略证书验证以继续获取信息
					})

					// 执行 TLS 握手
					if err := tlsConn.HandshakeContext(ctx); err != nil {
						tcpConn.Close()
						result.ErrorLog = fmt.Sprintf("TLS 握手失败: %v", err)
						result.StatusMsg = "TLS 握手失败"
						return nil, err
					}

					// 获取 SSL 证书信息（在 HTTP 请求之前）
					state := tlsConn.ConnectionState()
					if len(state.PeerCertificates) > 0 {
						cert := state.PeerCertificates[0]
						result.SSLExpiryDate = cert.NotAfter
						result.SSLCommonName = cert.Subject.CommonName
						result.SSLIssuerName = cert.Issuer.CommonName
						result.SSLSignatureAlg = cert.SignatureAlgorithm.String()
						result.SANList = strings.Join(cert.DNSNames, ",")

						// 证书链
						var chainBuilder strings.Builder
						for _, c := range state.PeerCertificates {
							pemBlock := &pem.Block{
								Type:  "CERTIFICATE",
								Bytes: c.Raw,
							}
							chainBuilder.WriteString(string(pem.EncodeToMemory(pemBlock)))
						}
						result.CertificateChain = chainBuilder.String()
					}

					result.StatusMsg = "TLS 握手成功"
					return tlsConn, nil
				}

				// HTTP 协议，使用普通 TCP 连接
				return dialer.DialContext(ctx, network, addr)
			},
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // 对于 HTTPS，我们已经手动处理证书
			},
		},
	}

	// --- 执行HTTP/HTTPS请求 ---
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		result.ErrorLog = fmt.Sprintf("创建请求失败: %v", err)
		result.StatusMsg = "请求创建失败"
		result.Duration = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		if result.ErrorLog == "" {
			result.ErrorLog = err.Error()
		}
		if result.StatusMsg == "未执行" {
			result.StatusMsg = "HTTP/HTTPS 请求失败"
		}
		return result
	}
	defer resp.Body.Close()

	// 处理 HTTP 状态
	result.HTTPStatus = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.IsSuccess = true
		if result.StatusMsg == "TLS 握手成功" {
			result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK (TLS 成功)", resp.StatusCode)
		} else {
			result.StatusMsg = fmt.Sprintf("HTTP 状态码: %d OK", resp.StatusCode)
		}
	} else {
		result.StatusMsg = fmt.Sprintf("HTTP 状态码错误: %d", resp.StatusCode)
	}

	return result
}
