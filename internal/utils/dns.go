package utils

import (
	"net"
	"strings"
)

// DNSValuesEqual 智能比较两个 DNS 记录值（支持 A、AAAA、CNAME）
func DNSValuesEqual(val1, val2 string) bool {
	// 1. 去除首尾空格
	v1 := strings.TrimSpace(val1)
	v2 := strings.TrimSpace(val2)

	if v1 == "" || v2 == "" {
		return v1 == v2
	}

	// 2. 尝试解析为 IP（IPv4 或 IPv6）
	ip1 := net.ParseIP(v1)
	ip2 := net.ParseIP(v2)

	if ip1 != nil && ip2 != nil {
		// 都是合法 IP（无论是 IPv4 还是 IPv6），用 Equal 比较
		return ip1.Equal(ip2)
	}

	// 3. 如果不是 IP，则当作域名处理（如 CNAME）
	//    域名不区分大小写，且末尾的 "." 可选（RFC 允许）
	domain1 := strings.ToLower(v1)
	domain2 := strings.ToLower(v2)

	// 去掉末尾的点（如果存在）
	domain1 = strings.TrimSuffix(domain1, ".")
	domain2 = strings.TrimSuffix(domain2, ".")

	return domain1 == domain2
}
