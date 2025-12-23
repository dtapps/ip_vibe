package ssl_helper

import (
	"context"
	"crypto"
	"encoding/json"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
)

// 参数
type SSLProcessParams struct {
	SSLRecordID string // SSL记录ID

	ConfigJSON json.RawMessage // 配置
	Domains    []string        // 域名列表
}

type SSLProcess interface {
	// 配置
	Configure(ctx context.Context, params *SSLProcessParams) error
	// 申请
	Apply(ctx context.Context, domains []string) (*certificate.Resource, error)
}

type LegoConfig struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *LegoConfig) GetEmail() string {
	return u.Email
}

func (u *LegoConfig) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *LegoConfig) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
