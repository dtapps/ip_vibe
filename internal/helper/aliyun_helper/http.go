package aliyun_helper

import (
	"net/http"

	"go.dtapp.net/library/contrib/http_log"
)

// HttpHandler 适配 github.com/alibabacloud-go/tea/tea 里的 HttpClient 接口
type HttpHandler struct {
	httpClient *http.Client
}

// NewHttpHandler 创建一个新的 HttpHandler 实例
func NewHttpHandler(httpTransport *http_log.LoggingRoundTripper) *HttpHandler {
	return &HttpHandler{
		httpClient: &http.Client{
			Transport: httpTransport,
		},
	}
}

// Call 实现了 dara.HttpClient 接口要求的方法
func (a *HttpHandler) Call(request *http.Request, transport *http.Transport) (*http.Response, error) {
	return a.httpClient.Do(request)
}
