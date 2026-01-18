package utils

import (
	"fmt"
	"net"
	"regexp"
	"strings"
	"vibe_ip/internal/ent/schema/schematype"

	"github.com/spf13/cast"
	"golang.org/x/net/publicsuffix"
)

// EqualDomainIgnoreCasePtr 忽略大小写比较域名是否相等
func EqualDomainIgnoreCasePtr(a, b *string) bool {
	if a == nil || b == nil {
		return false
	}
	return EqualDomainIgnoreCase(*a, *b)
}

// EqualIgnoreCase 忽略大小写比较两个字符串是否相等
// 泛型版本，适用于任何可转换为字符串的类型
func EqualIgnoreCase[T any](a, b T) bool {
	return EqualDomainIgnoreCase(cast.ToString(a), cast.ToString(b))
}

// EqualDomainIgnoreCase 忽略大小写比较域名是否相等
func EqualDomainIgnoreCase(a, b string) bool {
	// 去除前后空格
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	// 去除开头和结尾的 .
	a = strings.Trim(a, ".")
	b = strings.Trim(b, ".")
	return strings.EqualFold(a, b)
}

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
func GetRecordType(origin string) (string, error) {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return "", fmt.Errorf("输入不能为空")
	}

	// 判断是否为 IP 地址
	ip := net.ParseIP(origin)
	if ip != nil {
		if ip.To4() != nil {
			return "A", nil // IPv4
		}
		return "AAAA", nil // IPv6
	}

	// 判断是否为合法域名（简化版正则）
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)
	if domainRegex.MatchString(origin) {
		return "CNAME", nil
	}

	return "", fmt.Errorf("无效的输入，既不是 IP 也不是合法域名: %s", origin)
}

// GetFullDomain 组合子域名和主域名，返回完整域名
func GetFullDomain(subDomain string, mainDomain string) string {
	if subDomain == "@" {
		return mainDomain
	}
	return fmt.Sprintf("%s.%s", subDomain, mainDomain)
}

// GetDnsDomainParts 提取域名中的主域名和子域，同时判断记录类型
/**
	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName, recordType, err := utils.GetDnsDomainParts(subDomain, domainName, recordValue)
	if err != nil {
		return QueryDomainResponse{}, err
	}
**/
func GetDnsDomainParts(subDomain string, mainDomain string, originValue string) (full string, sub string, main string, recordType string, err error) {
	if subDomain == "" {
		subDomain = "@"
	}
	if originValue != "" {
		recordType, err = GetRecordType(originValue)
	}
	return fmt.Sprintf("%s.%s", subDomain, mainDomain), subDomain, mainDomain, recordType, err
}

// GetCdnDomainParts 提取域名中的主域名和子域，同时判断记录类型
/**
	// 提取域名中的主域名和子域，同时判断记录类型
	fullDomain, subDomain, domainName:= utils.GetCdnDomainParts(subDomain, domainName)
**/
func GetCdnDomainParts(subDomain string, mainDomain string) (full string, sub string, main string) {
	if subDomain == "" {
		subDomain = "@"
	}
	return fmt.Sprintf("%s.%s", subDomain, mainDomain), subDomain, mainDomain
}

// ParseDomain 解析域名
func ParseDomain(sD string, dn string) (fullDomain string, subDomain string, domainName string) {
	if sD == "" {
		sD = "@"
	}
	return fmt.Sprintf("%s.%s", sD, dn), sD, dn
}

// ParseCdnDomainType 解析域名类型
func ParseCdnDomainType(platform string, oV []string) (recordType string, originValues []string, err error) {
	switch platform {
	case schematype.ConfigPlatformTencentEO:
		if len(oV) >= 2 {
			recordType = "ORIGIN_GROUP"
		} else {
			recordType = "IP_DOMAIN"
		}
	case schematype.ConfigPlatformAliyunEsa:
		if len(oV) >= 2 {
			recordType = "CNAME"
		} else {
			recordType = "A/AAAA"
		}
	default:
		err = fmt.Errorf("平台 %s 暂不支持", platform)
	}
	return recordType, oV, err
}

// ParseDnsDomainType 解析域名类型
func ParseDnsDomainType(oV string) (recordType string, originValue string, err error) {
	recordType, err = GetRecordType(oV)
	return recordType, oV, err
}

// TrimDots 去除字符串首尾的点号
func TrimDotsPrt(s *string) string {
	if s == nil {
		return ""
	}
	return TrimDots(*s)
}

// TrimDots 去除字符串首尾的点号
func TrimDots(s string) string {
	return strings.Trim(s, ".")
}
