package cloudflare_helper

import (
	"fmt"

	"resty.dev/v3"
)

type Client struct {
	*resty.Client
}

// NewClient 创建请求客户端
func NewClient(api_key string) (*Client, error) {
	if api_key == "" {
		return nil, fmt.Errorf("check api_key")
	}

	client := resty.New().SetBaseURL("https://api.cloudflare.com/client/v4")
	client.SetResponseMiddlewares(
		Ensure2xxResponseMiddleware,       // 先调用，判断状态是不是请求成功
		resty.AutoParseResponseMiddleware, // 再调用，才能先判断状态码再解析
	)

	// 设置令牌
	client.SetAuthToken(api_key)

	return &Client{client}, nil
}

// WithDebug 开启调试模式
func (c *Client) WithDebug() *Client {
	c.EnableDebug()
	return c
}

// R 返回一个自定义的 Request，以便我们可以调用 SetBodyMap() SetBodyStruct() 解决因 body 顺序不同导致 SHA256 不一样的问题
func (c *Client) R() *Request {
	return &Request{c.Client.R().SetContentType("application/json")}
}
