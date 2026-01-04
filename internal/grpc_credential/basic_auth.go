package grpc_credential

import (
	"context"
	"log/slog"
	"vibe_ip/internal/core"
)

type BasicAuth struct {
	logger   *slog.Logger
	username string
	password string
}

func NewBasicAuth(logger *slog.Logger, username, password string) *BasicAuth {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	return &BasicAuth{
		logger:   logger,
		username: username,
		password: password,
	}
}

func (b *BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {

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

func (b *BasicAuth) RequireTransportSecurity() bool {
	return false // 若启用 TLS，建议返回 true
}
