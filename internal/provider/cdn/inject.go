package cdn

import (
	"sync"

	"go.dtapp.net/library/contrib/http_log"
)

var (
	cdnHttpTransport     *http_log.LoggingRoundTripper
	cdnHttpTransportOnce sync.Once
)

// 注入HTTP传输客户端
func InjectHttpTransport(httpTransport *http_log.LoggingRoundTripper) {
	cdnHttpTransportOnce.Do(func() {
		cdnHttpTransport = httpTransport
	})
}
