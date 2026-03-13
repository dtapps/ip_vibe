package dns

import (
	"sync"

	"go.dtapp.net/library/contrib/http_log"
)

var (
	dnsHttpTransport     *http_log.LoggingRoundTripper
	dnsHttpTransportOnce sync.Once
)

// 注入HTTP传输客户端
func InjectHttpTransport(httpTransport *http_log.LoggingRoundTripper) {
	dnsHttpTransportOnce.Do(func() {
		dnsHttpTransport = httpTransport
	})
}
