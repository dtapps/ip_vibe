package utils

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// ExtractDomainParts 提取域名中的主域名和子域
func ExtractDomainParts(input string) (full, sub, main string, err error) {
	// 清理输入：去除协议、路径、端口等
	host := input

	if strings.Contains(host, "/") {
		if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
			host = "https://" + host
		}
		parsed, err := publicsuffix.EffectiveTLDPlusOne(host)
		if err != nil {
			return "", "", "", err
		}
		host = parsed
	}

	// 去除端口
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// 使用 publicsuffix 提取注册域名（主域名）
	eTLD1, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", "", "", fmt.Errorf("无法提取主域名: %v", err)
	}

	full = host
	main = eTLD1 // 正确的主域名，如 a.com

	// 提取子域部分：去掉 main 之后剩下的前缀
	sub = strings.TrimSuffix(strings.TrimSuffix(full, main), ".")
	if sub == full {
		sub = "" // 没有子域
	}

	return full, sub, main, nil
}

// GetRecordType 判断输入是 IPv4、IPv6 还是域名，返回对应的 DNS 记录类型："A", "AAAA", "CNAME"
func GetRecordType(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	// 判断是否为 IP 地址
	ip := net.ParseIP(input)
	if ip != nil {
		if ip.To4() != nil {
			return "A", nil // IPv4
		}
		return "AAAA", nil // IPv6
	}

	// 判断是否为合法域名（简化版正则）
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if domainRegex.MatchString(input) {
		return "CNAME", nil
	}

	return "", fmt.Errorf("无效的输入，既不是 IP 也不是合法域名: %s", input)
}
