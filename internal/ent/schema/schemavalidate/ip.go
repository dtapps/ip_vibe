package schemavalidate

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// ValidateIPV4 校验是否为合法的 IPv4 地址
func ValidateIPV4(s string) error {
	if len(s) == 0 {
		return nil
	}
	if strings.Contains(s, ":") {
		return fmt.Errorf("IPv4 address should not contain colons")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s / %s", s, ip)
	}
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("invalid IPv4 address: %s / %s", s, ip)
	}
	return nil
}

// ValidateIPV6 校验是否为合法的 IPv6 地址
func ValidateIPV6(s string) error {
	if len(s) == 0 {
		return nil
	}
	if !strings.Contains(s, ":") {
		return fmt.Errorf("IPv6 address should contain colons")
	}
	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("invalid IP address: %s / %s", s, ip)
	}
	if ip == nil || ip.To4() != nil {
		return fmt.Errorf("invalid IPv6 address: %s / %s", s, ip)
	}
	return nil
}

// ValidateSubDomain 校验子域名部分
func ValidateSubDomain(s string) error {
	if len(s) == 0 {
		return nil
	}
	// 支持根域名解析 (@) 和 泛域名 (*)
	if s == "@" || s == "*" {
		return nil
	}
	// 子域名正则：允许字母、数字、下划线(部分服务商支持)和中划线，且支持多级子域名 (如 api.v1)
	// 结尾不能是点
	reg := regexp.MustCompile(`^([a-zA-Z0-9_]([a-zA-Z0-9-_]*[a-zA-Z0-9_])?\.)*[a-zA-Z0-9_]([a-zA-Z0-9-_]*[a-zA-Z0-9_])?$`)
	if !reg.MatchString(s) {
		return fmt.Errorf("invalid sub_domain format: %s", s)
	}
	return nil
}

// ValidateDomainName 校验是否为合法的根域名名称
func ValidateDomainName(s string) error {
	if len(s) == 0 {
		return nil
	}
	// 基本字符检查（不能包含协议、特殊字符或子域点数过多）
	if strings.Contains(s, "://") {
		return fmt.Errorf("domain name should not contain protocol (http/https)")
	}
	// 使用正则判断标准域名格式
	reg := regexp.MustCompile(`^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$`)
	if !reg.MatchString(strings.ToLower(s)) {
		return fmt.Errorf("invalid domain name format: %s", s)
	}
	return nil
}

// ValidateFullDomain 校验完整域名 (例如 www.example.com)
func ValidateFullDomain(s string) error {
	if len(s) == 0 {
		return nil
	}
	// 完整域名校验逻辑与 ValidateDomainName 类似，但通常用于存储拼接后的结果
	// 这里额外增加长度限制 (RFC 1035 标准)
	if len(s) > 253 {
		return fmt.Errorf("full_domain is too long (max 253 chars)")
	}
	reg := regexp.MustCompile(`^([a-z0-9*]([a-z0-9-_]*[a-z0-9])?\.)+[a-z]{2,}$`)
	if !reg.MatchString(strings.ToLower(s)) {
		return fmt.Errorf("invalid full_domain format: %s", s)
	}
	return nil
}
