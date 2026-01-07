package cdn_helper

import (
	"context"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/cdn/client"
	"go.dtapp.net/library/contrib/http_log"
)

type JdcloudCdnConfig struct {
	region     string // 区域
	AccessKey  string `json:"access_key"`  // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥密钥
	DomainName string `json:"domain_name"` // 主域名
}

type JdcloudCdnClient struct {
	config *JdcloudCdnConfig // 配置
	debug  bool              // 是否开启调试模式
	auth   *core.Credential
	client *client.CdnClient
}

// 京东云 内容分发网络 CDN
// https://www.jdcloud.com/cn/products/cdn
// https://docs.jdcloud.com/cn/cdn/api/overview
func NewJdcloudCdnClient(ctx context.Context, config *JdcloudCdnConfig) (*JdcloudCdnClient, error) {

	c := &JdcloudCdnClient{}
	c.config = config
	c.config.region = "cn-north-1"

	// 认证
	c.auth = core.NewCredentials(c.config.AccessKey, c.config.SecretKey)

	// 实例化
	c.client = client.NewCdnClient(c.auth)

	return c, nil
}

// 设置调试模式
func (c *JdcloudCdnClient) SetDebug(debug bool) *JdcloudCdnClient {
	c.debug = debug
	return c
}

// 设置HTTP日志记录器
func (c *JdcloudCdnClient) SetLog(loggingTransport *http_log.LoggingTransport) *JdcloudCdnClient {
	if loggingTransport == nil {
		return c
	}
	return c
}
