package core

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/miekg/dns"
)

const (
	// TypeA is the DNS type for A records.
	TypeA = "A"
	// TypeAAAA is the DNS type for AAAA records.
	TypeAAAA = "AAAA"
	// TypeCNAME is the DNS type for CNAME records.
	TypeCNAME = "CNAME"
)

// 返回 DNS 解析结果
type DNSResult struct {
	Type  string `json:"type,omitempty"`
	Value string `json:"value,omitempty"`
}

// resolveDNS 使用指定 DNS 服务器解析 A/AAAA/CNAME 记录
func resolveDNS(ctx context.Context, domain string, server string) (result []DNSResult, err error) {
	c := dns.Client{
		Timeout: 5 * time.Second,
	}
	m := dns.Msg{}

	// 查 A
	m.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	aR, _, err := c.Exchange(&m, server)
	if err != nil {
		return result, err
	}
	for _, ans := range aR.Answer {
		if a, ok := ans.(*dns.A); ok {
			result = append(result, DNSResult{
				Type:  TypeA,
				Value: a.A.String(),
			})
		}
	}

	// 查 AAAA
	m.SetQuestion(dns.Fqdn(domain), dns.TypeAAAA)
	aaaaR, _, err := c.Exchange(&m, server)
	if err != nil {
		return result, err
	}
	for _, ans := range aaaaR.Answer {
		if aaaa, ok := ans.(*dns.AAAA); ok {
			result = append(result, DNSResult{
				Type:  TypeAAAA,
				Value: aaaa.AAAA.String(),
			})
		}
	}

	// 查 CNAME
	m.SetQuestion(dns.Fqdn(domain), dns.TypeCNAME)
	cnameR, _, err := c.Exchange(&m, server)
	if err != nil {
		return result, err
	}
	for _, ans := range cnameR.Answer {
		if cname, ok := ans.(*dns.CNAME); ok {
			result = append(result, DNSResult{
				Type:  TypeCNAME,
				Value: cname.Target,
			})
		}
	}

	return result, nil
}

// joinHostPort 拼接主机名和端口号
func joinHostPort(host string, port int) string {
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}

// GetDNSServerList 获取 DNS 服务地址列表
func GetDNSServerList() (servers []string) {
	servers = []string{
		joinHostPort("119.29.29.29", 53),
		joinHostPort("2402:4e00::", 53),
		joinHostPort("223.5.5.5", 53),
		joinHostPort("2400:3200::1", 53),
		joinHostPort("180.76.76.76", 53),
		joinHostPort("2400:da00::6666", 53),
		joinHostPort("180.184.1.1", 53),
		joinHostPort("101.226.4.6", 53),
		joinHostPort("1.1.1.1", 53),
		joinHostPort("2606:4700:4700::1111", 53),
		joinHostPort("8.8.8.8", 53),
		joinHostPort("2001:4860:4860::8888", 53),
	}
	return servers
}

// GetDNSResultList 获取 DNS 解析结果列表
func GetDNSResultList(ctx context.Context, domain string, server string) (dnsResults []string, err error) {
	result, err := resolveDNS(ctx, domain, server)
	if err != nil {
		return dnsResults, err
	}
	for _, res := range result {
		dnsResults = append(dnsResults, res.Value)
	}
	return dnsResults, nil
}
