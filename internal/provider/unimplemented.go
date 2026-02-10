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
func (UnimplementedDNSProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (data *DNSProviderInternalData, err error) {
	return nil, ErrUnimplemented("DNSProvider", "Check")
}

// Update 更新 DNS 记录
func (UnimplementedDNSProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (err error) {
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
func (UnimplementedCDNProvider) Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (data *CDNProviderInternalData, err error) {
	return nil, ErrUnimplemented("CDNProvider", "Check")
}

// Update 更新 CDN 记录
func (UnimplementedCDNProvider) Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (err error) {
	return ErrUnimplemented("CDNProvider", "Update")
}

// ParseCustomConfig 解析 DNS 自定义配置
func (UnimplementedCDNProvider) ParseCustomConfig(configJSON json.RawMessage) (err error) {
	return ErrUnimplemented("CDNProvider", "ParseCustomConfig")
}

// ---------- TUN 模块 ----------

// UnimplementedTUNProvider 是一个“安全默认实现”，
// 所有方法都返回 Unimplemented 错误。
// 建议所有 TUNProvider 实现都嵌入它。
type UnimplementedTUNProvider struct{}

// Name 返回 TUN 提供程序的名称
func (UnimplementedTUNProvider) Name() string {
	return "Unimplemented"
}

// Check 检查 TUN 记录是否存在
func (UnimplementedTUNProvider) Check(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (data *TUNProviderInternalData, err error) {
	return nil, ErrUnimplemented("TUNProvider", "Check")
}

// Update 更新 TUN 记录
func (UnimplementedTUNProvider) Update(ctx context.Context, sub_domain string, domain_path string, serve_type string, serve_uri string) (err error) {
	return ErrUnimplemented("TUNProvider", "Update")
}

// ParseCustomConfig 解析 TUN 自定义配置
func (UnimplementedTUNProvider) ParseCustomConfig(configJSON json.RawMessage) (err error) {
	return ErrUnimplemented("TUNProvider", "ParseCustomConfig")
}

// ---------- TASK 模块 ----------

// UnimplementedTaskProvider 是一个“安全默认实现”，
// 所有方法都返回 Unimplemented 错误。
// 建议所有 TaskProvider 实现都嵌入它。
type UnimplementedTaskProvider struct{}

// Process 执行具体的任务逻辑
func (UnimplementedTaskProvider) Process(ctx context.Context, args json.RawMessage, meta TaskMetadata) (string, error) {
	return "", ErrUnimplemented("TaskProvider", "Process")
}
