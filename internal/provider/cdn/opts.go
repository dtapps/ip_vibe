package cdn

import (
	"go.dtapp.net/library/contrib/http_log"
)

// CdnOption 定义CDN配置选项函数类型
type CdnOption func(*cdnOptions)

// cdnOptions 内部CDN配置选项
type cdnOptions struct {
	httpTransport *http_log.LoggingRoundTripper
}

// WithHttpTransport 设置HTTP日志传输器
func WithHttpTransport(transport *http_log.LoggingRoundTripper) CdnOption {
	return func(o *cdnOptions) {
		o.httpTransport = transport
	}
}
