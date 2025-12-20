package cdn_helper

import (
	"context"

	"github.com/qiniu/go-sdk/v7/auth"
	"github.com/qiniu/go-sdk/v7/cdn"
)

type QiniuCdnConfig struct {
	AccessKey  string `json:"access_key"`  // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥Key
	DomainName string `json:"domain_name"` // 主域名
}

type QiniuCdnClient struct {
	config *QiniuCdnConfig // 配置
	debug  bool            // 是否开启调试模式
	auth   *auth.Credentials
	client *cdn.CdnManager
}

// 七牛 CDN
// https://www.qiniu.com/products/qcdn
func NewQiniuCdnClient(ctx context.Context, config *QiniuCdnConfig) (*QiniuCdnClient, error) {

	c := &QiniuCdnClient{}
	c.config = config

	// 密钥信息
	c.auth = auth.New(c.config.AccessKey, c.config.SecretKey)

	// 实例化要请求产品的client对象
	c.client = cdn.NewCdnManager(c.auth)

	return c, nil
}

// 设置调试模式
func (c *QiniuCdnClient) SetDebug(debug bool) *QiniuCdnClient {
	c.debug = debug
	return c
}
