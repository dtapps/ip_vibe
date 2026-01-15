package provider

import (
	"context"
	"encoding/json"
	"fmt"
)

// 辅助错误函数
func ErrUnimplemented(name string, method string) error {
	return fmt.Errorf("%s method %s not implemented", name, method)
}

// ---------- DNS 模块 ----------

// UnimplementedDNSProvider 是一个“安全默认实现”，
// 所有方法都返回 Unimplemented 错误。
// 建议所有 DNSProvider 实现都嵌入它。
type UnimplementedDNSProvider struct{}

// Name 返回 DNS 提供程序的名称
func (UnimplementedDNSProvider) Name() string {
	return "Unimplemented"
}

// Check 检查 DNS 记录是否存在
func (UnimplementedDNSProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, originValue string, err error) {
	return "", "", ErrUnimplemented("DNSProvider", "Check")
}

// Update 更新 DNS 记录
func (UnimplementedDNSProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValue string, recordType string) (err error) {
	return ErrUnimplemented("DNSProvider", "Update")
}

// ParseCustomConfig 解析 DNS 自定义配置
func (UnimplementedDNSProvider) ParseCustomConfig(configJSON json.RawMessage) (err error) {
	return ErrUnimplemented("DNSProvider", "ParseCustomConfig")
}

// ---------- CDN 模块 ----------

// UnimplementedCDNProvider 是一个“安全默认实现”，
// 所有方法都返回 Unimplemented 错误。
// 建议所有 CDNProvider 实现都嵌入它。
type UnimplementedCDNProvider struct{}

// Name 返回 CDN 提供程序的名称
func (UnimplementedCDNProvider) Name() string {
	return "Unimplemented"
}

// Check 检查 CDN 记录是否存在
func (UnimplementedCDNProvider) Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, cdnDomain string, originValues []string, err error) {
	return "", "", nil, ErrUnimplemented("CDNProvider", "Check")
}

// Update 更新 CDN 记录
func (UnimplementedCDNProvider) Update(ctx context.Context, recordID string, domain string, subdomain string, originValues []string, recordType string) (err error) {
	return ErrUnimplemented("CDNProvider", "Update")
}

// ParseCustomConfig 解析 DNS 自定义配置
func (UnimplementedCDNProvider) ParseCustomConfig(configJSON json.RawMessage) (err error) {
	return ErrUnimplemented("CDNProvider", "ParseCustomConfig")
}
