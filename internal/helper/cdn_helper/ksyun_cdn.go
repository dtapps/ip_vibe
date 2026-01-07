package cdn_helper

import (
	"context"

	cdn "github.com/kingsoftcloud/sdk-go/v2/ksyun/client/cdn/v3"
	"github.com/kingsoftcloud/sdk-go/v2/ksyun/common"
	"github.com/kingsoftcloud/sdk-go/v2/ksyun/common/profile"
)

type KsyunCdnConfig struct {
	region     string
	AccessKey  string `json:"access_key"`  // 密钥ID
	SecretKey  string `json:"secret_key"`  // 密钥Key
	DomainName string `json:"domain_name"` // 主域名
}

type KsyunCdnClient struct {
	config  *KsyunCdnConfig // 配置
	debug   bool            // 是否开启调试模式
	auth    *common.Credential
	profile *profile.ClientProfile
	client  *cdn.Client
}

// 金山云内容分发网络
// https://www.ksyun.com/nv/product/CDN.html
// https://apiexplorer.ksyun.com/#/api/0/1031
func NewKsyunCdnClient(ctx context.Context, config *KsyunCdnConfig) (*KsyunCdnClient, error) {

	var err error

	c := &KsyunCdnClient{}
	c.config = config
	c.config.region = "cdn_cn"

	// 密钥信息
	c.auth = common.NewCredential(c.config.AccessKey, c.config.SecretKey)

	// 客户端配置
	c.profile = profile.NewClientProfile()
	c.profile.HttpProfile.ReqMethod = "POST|GET"
	c.profile.HttpProfile.ReqTimeout = 60
	c.profile.HttpProfile.Endpoint = "cdn.api.ksyun.com"

	// 实例化要请求产品的client对象
	c.client, err = cdn.NewClient(c.auth, c.config.region, c.profile)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// 设置调试模式
func (c *KsyunCdnClient) SetDebug(debug bool) *KsyunCdnClient {
	c.debug = debug
	return c
}
