package cdn_helper

import (
	"context"

	"github.com/baidubce/bce-sdk-go/services/cdn"
)

type BaiduCdnConfig struct {
	AccessKeyID     string `json:"access_key_id"`     // 密钥ID
	SecretAccessKey string `json:"secret_access_key"` // 密钥Key
	DomainName      string `json:"domain_name"`       // 主域名
}

type BaiduCdnClient struct {
	config *BaiduCdnConfig // 配置
	debug  bool            // 是否开启调试模式
	client *cdn.Client
}

// 百度云 内容分发网络 CDN
// https://cloud.baidu.com/product/CDN.html
// https://cloud.baidu.com/doc/CDN/s/1lq6g08z6
func NewBaiduCdnClient(ctx context.Context, config *BaiduCdnConfig) (*BaiduCdnClient, error) {

	var err error

	c := &BaiduCdnClient{}
	c.config = config

	c.client, err = cdn.NewClient(c.config.AccessKeyID, c.config.SecretAccessKey, "cdn.baidubce.com")
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *BaiduCdnClient) SetDebug(debug bool) *BaiduCdnClient {
	c.debug = debug
	return c
}
