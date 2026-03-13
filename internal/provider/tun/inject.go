package tun

import (
	"sync"

	"go.dtapp.net/library/contrib/http_log"
)

var (
	tunHttpTransport     *http_log.LoggingRoundTripper
	tunHttpTransportOnce sync.Once
)

// 注入HTTP传输客户端
func InjectHttpTransport(httpTransport *http_log.LoggingRoundTripper) {
	tunHttpTransportOnce.Do(func() {
		tunHttpTransport = httpTransport
	})
}
