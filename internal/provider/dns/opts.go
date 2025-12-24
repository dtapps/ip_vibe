package dns

import (
	"go.dtapp.net/library/contrib/http_log"
)

// DnsOption 定义DNS配置选项函数类型
type DnsOption func(*dnsOptions)

// dnsOptions 内部DNS配置选项
type dnsOptions struct {
	httpTransport *http_log.LoggingTransport
}

// WithHttpTransport 设置HTTP日志传输器
func WithHttpTransport(transport *http_log.LoggingTransport) DnsOption {
	return func(o *dnsOptions) {
		o.httpTransport = transport
	}
}
