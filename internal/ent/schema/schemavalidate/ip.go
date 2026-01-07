package schemavalidate

import (
	"fmt"
	"net"
	"strings"
)

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
