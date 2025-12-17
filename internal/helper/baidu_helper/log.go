package baidu_helper

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// LogData 表示每次请求/响应的日志数据
type LogData struct {
	GoVersion       string
	Hostname        string
	Method          string
	URL             string
	RequestHeaders  http.Header
	RequestBody     json.RawMessage
	StatusCode      int
	ResponseHeaders http.Header
	ResponseBody    json.RawMessage
	ElapseTime      int64
	IsError         bool
}

// 定义拦截器
type LoggingTransport struct {
	Proxied http.RoundTripper
}

func (l *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// 记录请求
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}
	log.Printf("[DNS REQ] %s %s | Body: %s", req.Method, req.URL.String(), string(reqBody))

	// 执行
	resp, err := l.Proxied.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// 记录响应
	var respBody []byte
	if resp.Body != nil {
		respBody, _ = io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewBuffer(respBody))
	}
	log.Printf("[DNS RESP] Status: %s | Body: %s", resp.Status, string(respBody))

	return resp, nil
}
