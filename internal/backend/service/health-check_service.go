package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/sync/errgroup"
)

type HealthCheckResult struct {
	TargetDomain string `json:"target_domain,omitempty"` // 检测的域名（用于 ToPairsForDomain）
	TargetIP     string `json:"target_ip,omitempty"`     // 检测的 IP（用于 ToPairsForIP）
	HTTPPort     int    `json:"http_port,omitempty"`     // 显式记录端口
	HTTPSPort    int    `json:"https_port,omitempty"`    // 显式记录端口

	DNSUsed     string   `json:"dns_used,omitempty"`     // 实际使用的 DNS 服务器
	ResolvedIPs []string `json:"resolved_ips,omitempty"` // 解析出的 IP 地址

	HTTPStatus    int    `json:"http_status,omitempty"` // HTTP 状态码
	HTTPErr       string `json:"http_err,omitempty"`    // HTTP 错误信息
	HTTPRTT       int64  `json:"http_rtt_ms,omitempty"` // HTTP 响应时间（毫秒）
	IsHTTPHealthy bool   `json:"is_http_healthy"`       // 是否 HTTP 健康

	HTTPTCPErr       string `json:"http_tcp_err,omitempty"`    // HTTP TCP 错误信息
	HTTPTCPRTT       int64  `json:"http_tcp_rtt_ms,omitempty"` // HTTP TCP 响应时间（毫秒）
	IsHTTPTCPHealthy bool   `json:"is_http_tcp_healthy"`       // 是否 HTTP TCP 健康

	HTTPUDPErr       string `json:"http_udp_err,omitempty"`    // HTTP UDP 错误信息
	HTTPUDPRTT       int64  `json:"http_udp_rtt_ms,omitempty"` // HTTP UDP 响应时间（毫秒）
	IsHTTPUDPHealthy bool   `json:"is_http_udp_healthy"`       // 是否 HTTP UDP 健康

	HTTPSStatus    int    `json:"https_status,omitempty"` // HTTPS 状态码
	HTTPSErr       string `json:"https_err,omitempty"`    // HTTPS 错误信息
	HTTPSRTT       int64  `json:"https_rtt_ms,omitempty"` // HTTPS 响应时间（毫秒）
	IsHTTPSHealthy bool   `json:"is_https_healthy"`       // 是否 HTTPS 健康

	HTTPSTCPErr       string `json:"https_tcp_err,omitempty"`    // HTTPS TCP 错误信息
	HTTPSTCPRTT       int64  `json:"https_tcp_rtt_ms,omitempty"` // HTTPS TCP 响应时间（毫秒）
	IsHTTPSTCPHealthy bool   `json:"is_https_tcp_healthy"`       // 是否 HTTPS TCP 健康

	HTTPSUDPErr       string `json:"https_udp_err,omitempty"`    // HTTPS UDP 错误信息
	HTTPSUDPRTT       int64  `json:"https_udp_rtt_ms,omitempty"` // HTTPS UDP 响应时间（毫秒）
	IsHTTPSUDPHealthy bool   `json:"is_https_udp_healthy"`       // 是否 HTTPS UDP 健康
}

// CheckDomainHealth 检测域名是否可通（含 DNS + HTTP + HTTPS）
func CheckDomainHealth(ctx context.Context, domain string, httpPort, httpsPort int, dnsServers []string) (*HealthCheckResult, error) {
	if len(dnsServers) == 0 {
		dnsServers = []string{
			JoinHostPort("119.29.29.29", 53),
			JoinHostPort("2402:4e00::", 53),
			JoinHostPort("223.5.5.5", 53),
			JoinHostPort("2400:3200::1", 53),
			JoinHostPort("180.76.76.76", 53),
			JoinHostPort("2400:da00::6666", 53),
			JoinHostPort("180.184.1.1", 53),
			JoinHostPort("101.226.4.6", 53),
			JoinHostPort("1.1.1.1", 53),
			JoinHostPort("2606:4700:4700::1111", 53),
			JoinHostPort("8.8.8.8", 53),
			JoinHostPort("2001:4860:4860::8888", 53),
		}
	}

	result := &HealthCheckResult{
		TargetDomain: domain,
		HTTPPort:     httpPort,
		HTTPSPort:    httpsPort,
	}

	// DNS 解析（尝试多个 DNS）
	var resolvedIPs []string
	var usedDNS string

	// 用于快速通知其他 goroutine 停止
	found := make(chan struct{}, 1)

	// 保护 resolvedIPs 和 usedDNS 的写入
	var mu sync.Mutex

	// 并发处理DNS解析
	dnsEg, dnsCtx := errgroup.WithContext(ctx)
	dnsEg.SetLimit(len(dnsServers))

	for _, server := range dnsServers {
		server := server // 闭包捕获
		dnsEg.Go(func() error {
			// 检查是否已有结果，避免无效工作
			select {
			case <-found:
				return nil // 已有成功结果，直接退出
			default:
			}

			ips, err := resolveDNS(dnsCtx, domain, server)
			if err == nil && len(ips) > 0 {
				mu.Lock()
				// 再次检查，防止多个 goroutine 同时成功
				if len(resolvedIPs) == 0 {
					resolvedIPs = ips
					usedDNS = server
					select {
					case found <- struct{}{}: // 发出完成信号
					default:
					}
				}
				mu.Unlock()
			}
			return nil
		})
	}

	// 等待 并发处理DNS解析 完成
	_ = dnsEg.Wait()

	result.DNSUsed = usedDNS
	result.ResolvedIPs = resolvedIPs

	// 并发处理
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(2) // 最多 2 个并发

	// HTTP 检测（如果端口 > 0）
	if httpPort > 0 {
		eg.Go(func() error {
			httpStatus, rtt, err := checkHTTP(egCtx, fmt.Sprintf("http://%s:%d", domain, httpPort), 10*time.Second)
			result.HTTPStatus = httpStatus
			result.HTTPRTT = rtt
			if err != nil {
				result.HTTPErr = err.Error()
			} else {
				result.IsHTTPHealthy = httpStatus >= 200 && httpStatus < 400
			}
			return nil
		})
	}

	// HTTPS 检测（如果端口 > 0）
	if httpsPort > 0 {
		eg.Go(func() error {
			httpsStatus, rtt, err := checkHTTP(egCtx, fmt.Sprintf("https://%s:%d", domain, httpsPort), 10*time.Second)
			result.HTTPSStatus = httpsStatus
			result.HTTPSRTT = rtt
			if err != nil {
				result.HTTPSErr = err.Error()
			} else {
				result.IsHTTPSHealthy = httpsStatus >= 200 && httpsStatus < 400
			}
			return nil
		})
	}

	// 等待 并发处理 完成
	_ = eg.Wait()

	return result, nil
}

// CheckIPHealth 检测 IP 是否可通过 HTTP/HTTPS 访问
func CheckIPHealth(ctx context.Context, ip string, httpPort, httpsPort int) (*HealthCheckResult, error) {
	result := &HealthCheckResult{
		TargetIP:  ip,
		HTTPPort:  httpPort,
		HTTPSPort: httpsPort,
	}

	// 并发处理
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(6) // 最多 6 个并发

	// HTTP 检测（如果端口 > 0）
	if httpPort > 0 {
		eg.Go(func() error {
			httpStatus, rtt, err := checkHTTP(egCtx, fmt.Sprintf("http://%s", JoinHostPort(ip, httpPort)), 10*time.Second)
			result.HTTPStatus = httpStatus
			result.HTTPRTT = rtt
			if err != nil {
				result.HTTPErr = err.Error()
			} else {
				result.IsHTTPHealthy = httpStatus >= 200 && httpStatus < 400
			}
			return nil
		})
	}
	// HTTP TCP 检测（如果端口 > 0）
	if httpPort > 0 {
		eg.Go(func() error {
			ok, rtt, err := checkTCP(egCtx, JoinHostPort(ip, httpPort), 10*time.Second)
			result.IsHTTPTCPHealthy = ok
			result.HTTPTCPRTT = rtt
			if err != nil {
				result.HTTPTCPErr = err.Error()
			}
			return nil
		})
	}
	// HTTP UDP 检测（如果端口 > 0）
	if httpPort > 0 {
		eg.Go(func() error {
			ok, rtt, err := checkUDP(egCtx, JoinHostPort(ip, httpPort), 10*time.Second)
			result.IsHTTPUDPHealthy = ok
			result.HTTPUDPRTT = rtt
			if err != nil {
				result.HTTPUDPErr = err.Error()
			}
			return nil
		})
	}

	// HTTPS 检测（如果端口 > 0）
	if httpsPort > 0 {
		eg.Go(func() error {
			httpsStatus, rtt, err := checkHTTP(egCtx, fmt.Sprintf("https://%s", JoinHostPort(ip, httpsPort)), 10*time.Second)
			result.HTTPSStatus = httpsStatus
			result.HTTPSRTT = rtt
			if err != nil {
				result.HTTPSErr = err.Error()
			} else {
				result.IsHTTPSHealthy = httpsStatus >= 200 && httpsStatus < 400
			}
			return nil
		})
	}
	// HTTPS TCP 检测（如果端口 > 0）
	if httpsPort > 0 {
		eg.Go(func() error {
			ok, rtt, err := checkTCP(egCtx, JoinHostPort(ip, httpsPort), 10*time.Second)
			result.IsHTTPSTCPHealthy = ok
			result.HTTPSTCPRTT = rtt
			if err != nil {
				result.HTTPSTCPErr = err.Error()
			}
			return nil
		})
	}
	// HTTPS UDP 检测（如果端口 > 0）
	if httpsPort > 0 {
		eg.Go(func() error {
			ok, rtt, err := checkUDP(egCtx, JoinHostPort(ip, httpsPort), 10*time.Second)
			result.IsHTTPSUDPHealthy = ok
			result.HTTPSUDPRTT = rtt
			if err != nil {
				result.HTTPSUDPErr = err.Error()
			}
			return nil
		})
	}

	// 等待 并发处理 完成
	_ = eg.Wait()

	return result, nil
}

// resolveDNS 使用指定 DNS 服务器解析 A/AAAA 记录
func resolveDNS(ctx context.Context, domain string, server string) (ips []string, err error) {
	c := dns.Client{Timeout: 5 * time.Second}
	m := dns.Msg{}

	// 查 A
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	r, _, err := c.Exchange(&m, server)
	if err != nil {
		return nil, err
	}
	for _, ans := range r.Answer {
		if a, ok := ans.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}

	// 查 AAAA
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

	// 查 CNAME
	m.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	r, _, err = c.Exchange(&m, server)
	if err != nil {
		return nil, err
	}
	for _, ans := range r.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			ips = append(ips, cname.Target)
		}
	}

	return ips, nil
}

// checkHTTP 发起 HTTP(S) 请求并返回状态码 + 耗时
func checkHTTP(ctx context.Context, url string, timeout time.Duration) (int, int64, error) {
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

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	req.Header.Set("User-Agent", "IPVibe-HealthChecker/1.0")
	resp, err := client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return 0, duration, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, duration, nil
}

// checkTCP 检测 TCP 连通性（返回是否成功 + 耗时）
func checkTCP(ctx context.Context, address string, timeout time.Duration) (bool, int64, error) {
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0, err
	}
	conn.Close()

	return true, duration, nil
}

// checkUDP 尝试发送 UDP 包（无响应验证，仅测是否能发）
func checkUDP(ctx context.Context, address string, timeout time.Duration) (bool, int64, error) {
	start := time.Now()

	// 解析地址
	serverAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return false, 0, err
	}

	// 创建 UDP 连接
	conn, err := net.DialUDP("udp", nil, serverAddr)
	duration := time.Since(start).Milliseconds()
	if err != nil {
		return false, 0, err
	}
	defer conn.Close()

	// 设置读写超时
	writeDeadline := time.Now().Add(timeout)
	conn.SetWriteDeadline(writeDeadline)

	// 发送探测包（1 字节）
	testData := []byte{0x00}
	n, writeErr := conn.Write(testData)
	if writeErr != nil {
		return false, 0, writeErr
	}
	if n != len(testData) {
		return false, 0, fmt.Errorf("预期 %d bytes, 写了 %d", len(testData), n)
	}

	// UDP 无通用响应，这里仅认为“能发包”就算通
	return true, duration, nil
}

type ToPairsStruct struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// ToPairsForDomain 返回域名检测关注的字段
func (r *HealthCheckResult) ToPairsForDomain() (pairs []ToPairsStruct) {

	// 检测的域名
	if r.TargetDomain != "" {
		pairs = append(pairs, ToPairsStruct{Key: "检测的域名", Value: r.TargetDomain})
	}

	// 端口信息
	if r.HTTPPort > 0 {
		pairs = append(pairs, ToPairsStruct{Key: "HTTP 端口", Value: r.HTTPPort})
	}
	if r.HTTPSPort > 0 {
		pairs = append(pairs, ToPairsStruct{Key: "HTTPS 端口", Value: r.HTTPSPort})
	}

	if r.DNSUsed != "" {
		pairs = append(pairs, ToPairsStruct{Key: "使用的DNS服务器", Value: r.DNSUsed})
	}
	if len(r.ResolvedIPs) > 0 {
		pairs = append(pairs, ToPairsStruct{Key: "解析出的IP地址", Value: strings.Join(r.ResolvedIPs, ", ")})
	}

	// HTTP/HTTPS 部分保持不变
	pairs = addHTTPHTTPS(pairs, r)

	return pairs
}

// ToPairsForIP 返回 IP 检测关注的字段
func (r *HealthCheckResult) ToPairsForIP() (pairs []ToPairsStruct) {

	// 检测的 IP
	if r.TargetIP != "" {
		pairs = append(pairs, ToPairsStruct{Key: "检测的IP地址", Value: r.TargetIP})
	}

	// 端口信息
	if r.HTTPPort > 0 {
		pairs = append(pairs, ToPairsStruct{Key: "HTTP 端口", Value: r.HTTPPort})
	}
	if r.HTTPSPort > 0 {
		pairs = append(pairs, ToPairsStruct{Key: "HTTPS 端口", Value: r.HTTPSPort})
	}

	// HTTP/HTTPS URL 字段
	pairs = addHTTPHTTPS(pairs, r)

	// HTTP/HTTPS TCP/UDP 字段
	pairs = addHTTPHTTPSTCPUDP(pairs, r)

	return pairs
}

// 添加 HTTP/HTTPS URL 字段
func addHTTPHTTPS(pairs []ToPairsStruct, r *HealthCheckResult) []ToPairsStruct {
	if r.HTTPPort > 0 {
		if r.HTTPErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP 错误", Value: r.HTTPErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP 是否通", Value: boolToStatus(r.IsHTTPHealthy)})
			pairs = append(pairs, ToPairsStruct{Key: "HTTP 状态码", Value: r.HTTPStatus})
			if r.HTTPRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTP 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPRTT)})
			}
		}
	}

	if r.HTTPSPort > 0 {
		if r.HTTPSErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS 错误", Value: r.HTTPSErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS 是否通", Value: boolToStatus(r.IsHTTPSHealthy)})
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS 状态码", Value: r.HTTPSStatus})
			if r.HTTPSRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTPS 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPSRTT)})
			}
		}
	}

	return pairs
}

// 添加 HTTP/HTTPS TCP/UDP 字段
func addHTTPHTTPSTCPUDP(pairs []ToPairsStruct, r *HealthCheckResult) []ToPairsStruct {
	if r.HTTPPort > 0 {
		if r.HTTPTCPErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP TCP 错误", Value: r.HTTPTCPErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP TCP 是否通", Value: boolToStatus(r.IsHTTPTCPHealthy)})
			if r.HTTPTCPRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTP TCP 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPTCPRTT)})
			}
		}

		if r.HTTPUDPErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP UDP 错误", Value: r.HTTPUDPErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTP UDP 是否通", Value: boolToStatus(r.IsHTTPUDPHealthy)})
			if r.HTTPUDPRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTP UDP 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPUDPRTT)})
			}
		}
	}

	if r.HTTPSPort > 0 {
		if r.HTTPSTCPErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS TCP 错误", Value: r.HTTPSTCPErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS TCP 是否通", Value: boolToStatus(r.IsHTTPSTCPHealthy)})
			if r.HTTPSTCPRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTPS TCP 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPSTCPRTT)})
			}
		}

		if r.HTTPSUDPErr != "" {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS UDP 错误", Value: r.HTTPSUDPErr})
		} else {
			pairs = append(pairs, ToPairsStruct{Key: "HTTPS UDP 是否通", Value: boolToStatus(r.IsHTTPSUDPHealthy)})
			if r.HTTPSUDPRTT > 0 {
				pairs = append(pairs, ToPairsStruct{Key: "HTTPS UDP 响应时间", Value: fmt.Sprintf("%d ms", r.HTTPSUDPRTT)})
			}
		}
	}

	return pairs
}

func boolToStatus(ok bool) string {
	if ok {
		return "✅ 正常"
	}
	return "❌ 异常"
}

// 拼接主机名和端口号
func JoinHostPort(host string, port int) string {
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}
