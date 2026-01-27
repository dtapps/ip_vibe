package ssl

import (
	"go.dtapp.net/library/contrib/http_log"
)

// Option 定义SSL配置选项函数类型
type Option func(*Options)

// Options 内部SSL配置选项
type Options struct {
	httpTransport *http_log.LoggingRoundTripper
}

// WithHttpTransport 设置HTTP日志传输器
func WithHttpTransport(transport *http_log.LoggingRoundTripper) Option {
	return func(o *Options) {
		o.httpTransport = transport
	}
}
