package grpc_credential

import (
	"context"
	"encoding/base64"
	"log/slog"
)

type BasicAuth struct {
	logger   *slog.Logger
	username string
	password string
}

func NewBasicAuth(logger *slog.Logger, username, password string) *BasicAuth {
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

	cred := "Basic " + base64.StdEncoding.EncodeToString([]byte(b.username+":"+b.password))
	return map[string]string{"authorization": cred}, nil
}

func (b *BasicAuth) RequireTransportSecurity() bool {
	return false // 若启用 TLS，建议返回 true
}
