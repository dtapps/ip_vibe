package grpc_credential

import (
	"context"
	"log/slog"
	"vibe_ip/internal/core"
)

// ClientBasicAuth 实现了 grpc.PerRPCCredentials 接口
// 用于客户端在发起请求时自动注入 HTTP Basic Auth 头
type ClientBasicAuth struct {
	logger   *slog.Logger
	username string
	password string
}

// NewClientBasicAuth 创建一个新的客户端基础认证凭证
func NewClientBasicAuth(logger *slog.Logger, username, password string) *ClientBasicAuth {
	return &ClientBasicAuth{
		logger:   logger,
		username: username,
		password: password,
	}
}

func (b *ClientBasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

	// 如果用户名或密码为空就不校验
	if b.username == "" || b.password == "" {
		return nil, nil
	}

	// 获取基本认证 token
	token := core.GetBasicAuthHeader(core.GetBasicAuthToken(b.username, b.password))

	return map[string]string{
		"authorization": token,
	}, nil
}

func (b *ClientBasicAuth) RequireTransportSecurity() bool {
	return false // 若启用 TLS，建议返回 true
}
