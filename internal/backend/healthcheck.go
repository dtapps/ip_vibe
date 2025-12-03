package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

type HealthCheckResult struct {
	DNSUsed        string   `json:"dns_used,omitempty"`     // 实际使用的 DNS 服务器
	ResolvedIPs    []string `json:"resolved_ips,omitempty"` // 解析出的 IP 地址
	HTTPStatus     int      `json:"http_status,omitempty"`  // HTTP 状态码
	HTTPSStatus    int      `json:"https_status,omitempty"` // HTTPS 状态码
	HTTPErr        string   `json:"http_err,omitempty"`     // HTTP 错误信息
	HTTPSErr       string   `json:"https_err,omitempty"`    // HTTPS 错误信息
	HTTPRTT        int64    `json:"http_rtt_ms,omitempty"`  // HTTP 响应时间（毫秒）
	HTTPSRTT       int64    `json:"https_rtt_ms,omitempty"` // HTTPS 响应时间（毫秒）
	IsHTTPHealthy  bool     `json:"is_http_healthy"`        // 是否 HTTP 健康
	IsHTTPSHealthy bool     `json:"is_https_healthy"`       // 是否 HTTPS 健康
}

// CheckDomainHealth 检测域名是否可通（含 DNS + HTTP + HTTPS）
func CheckDomainHealth(ctx context.Context, domain string, httpPort, httpsPort int, dnsServers []string) (*HealthCheckResult, error) {
	if len(dnsServers) == 0 {
		dnsServers = []string{"119.29.29.29:53", "223.5.5.5:53", "8.8.8.8:53", "1.1.1.1:53"}
	}

	result := &HealthCheckResult{}

	// 1. DNS 解析（尝试多个 DNS）
	var resolvedIPs []string
	var usedDNS string
	for _, server := range dnsServers {
		ips, err := resolveDNS(ctx, domain, server)
		if err == nil && len(ips) > 0 {
			resolvedIPs = ips
			usedDNS = server
			break
		}
	}
	result.DNSUsed = usedDNS
	result.ResolvedIPs = resolvedIPs

	// 2. HTTP 检测（如果端口 > 0）
	if httpPort > 0 {
		httpStatus, rtt, err := checkHTTP(ctx, fmt.Sprintf("http://%s:%d", domain, httpPort))
		result.HTTPStatus = httpStatus
		result.HTTPRTT = rtt.Milliseconds()
		if err != nil {
			result.HTTPErr = err.Error()
		} else {
			result.IsHTTPHealthy = httpStatus >= 200 && httpStatus < 400
		}
	}

	// 3. HTTPS 检测（如果端口 > 0）
	if httpsPort > 0 {
		httpsStatus, rtt, err := checkHTTP(ctx, fmt.Sprintf("https://%s:%d", domain, httpsPort))
		result.HTTPSStatus = httpsStatus
		result.HTTPSRTT = rtt.Milliseconds()
		if err != nil {
			result.HTTPSErr = err.Error()
		} else {
			result.IsHTTPSHealthy = httpsStatus >= 200 && httpsStatus < 400
		}
	}

	return result, nil
}

// CheckIPHealth 检测 IP 是否可通过 HTTP/HTTPS 访问
func CheckIPHealth(ctx context.Context, ip string, httpPort, httpsPort int, expectedIPs []string) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		ResolvedIPs: []string{ip}, // 虚拟 DNS 结果
	}

	// 可选：验证 IP 是否在预期列表中
	if len(expectedIPs) > 0 {
		found := false
		for _, eip := range expectedIPs {
			if eip == ip {
				found = true
				break
			}
		}
		if !found {
			result.HTTPErr = "IP 不在预期列表中"
			result.HTTPSErr = "IP 不在预期列表中"
			return result, nil
		}
	}

	// 处理 IPv6 地址（HTTP 需方括号）
	ipStr := ip
	if strings.Contains(ip, ":") && !strings.HasPrefix(ip, "[") {
		ipStr = "[" + ip + "]"
	}

	// HTTP 检测
	if httpPort > 0 {
		status, rtt, err := checkHTTP(ctx, fmt.Sprintf("http://%s:%d", ipStr, httpPort))
		result.HTTPStatus = status
		result.HTTPRTT = rtt.Milliseconds()
		if err != nil {
			result.HTTPErr = err.Error()
		} else {
			result.IsHTTPHealthy = status >= 200 && status < 400
		}
	}

	// HTTPS 检测
	if httpsPort > 0 {
		status, rtt, err := checkHTTP(ctx, fmt.Sprintf("https://%s:%d", ipStr, httpsPort))
		result.HTTPSStatus = status
		result.HTTPSRTT = rtt.Milliseconds()
		if err != nil {
			result.HTTPSErr = err.Error()
		} else {
			result.IsHTTPSHealthy = status >= 200 && status < 400
		}
	}

	return result, nil
}

// resolveDNS 使用指定 DNS 服务器解析 A/AAAA 记录
func resolveDNS(ctx context.Context, domain string, server string) ([]string, error) {
	c := dns.Client{Timeout: 5 * time.Second}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	r, _, err := c.Exchange(&m, server)
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	// 如果没 A 记录，查 AAAA
	if len(ips) == 0 {
		m.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)
		r, _, err = c.Exchange(&m, server)
		if err != nil {
			return nil, err
		}
		for _, ans := range r.Answer {
			if aaaa, ok := ans.(*dns.AAAA); ok {
				ips = append(ips, aaaa.AAAA.String())
			}
		}
	}
	return ips, nil
}

// checkHTTP 发起 HTTP(S) 请求并返回状态码 + 耗时
func checkHTTP(ctx context.Context, url string) (int, time.Duration, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略证书错误（CDN 可能用自签）
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "IPVibe-HealthChecker/1.0")
	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return 0, duration, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, duration, nil
}

func (r *HealthCheckResult) ToPairs() []map[string]string {
	var pairs []map[string]string

	if r.DNSUsed != "" {
		pairs = append(pairs, map[string]string{
			"Key":   "使用的 DNS 服务器",
			"Value": r.DNSUsed,
		})
	}

	if len(r.ResolvedIPs) > 0 {
		pairs = append(pairs, map[string]string{
			"Key":   "解析出的 IP 地址",
			"Value": strings.Join(r.ResolvedIPs, ", "),
		})
	}

	if r.HTTPStatus > 0 {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTP 状态码",
			"Value": strconv.Itoa(r.HTTPStatus),
		})
	}
	if r.HTTPErr != "" {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTP 错误",
			"Value": r.HTTPErr,
		})
	}
	if r.HTTPRTT > 0 {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTP 响应时间（ms）",
			"Value": strconv.FormatInt(r.HTTPRTT, 10),
		})
	}

	if r.HTTPSStatus > 0 {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTPS 状态码",
			"Value": strconv.Itoa(r.HTTPSStatus),
		})
	}
	if r.HTTPSErr != "" {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTPS 错误",
			"Value": r.HTTPSErr,
		})
	}
	if r.HTTPSRTT > 0 {
		pairs = append(pairs, map[string]string{
			"Key":   "HTTPS 响应时间（ms）",
			"Value": strconv.FormatInt(r.HTTPSRTT, 10),
		})
	}

	pairs = append(pairs, map[string]string{
		"Key":   "HTTP 是否健康",
		"Value": boolToStatus(r.IsHTTPHealthy),
	})
	pairs = append(pairs, map[string]string{
		"Key":   "HTTPS 是否健康",
		"Value": boolToStatus(r.IsHTTPSHealthy),
	})

	return pairs
}

func boolToStatus(ok bool) string {
	if ok {
		return "✅ 正常"
	}
	return "❌ 异常"
}
