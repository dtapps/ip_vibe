package grpc_credential

import (
	"context"
	"encoding/base64"
)

type BasicAuth struct {
	Username string
	Password string
}

func (b *BasicAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	cred := "Basic " + base64.StdEncoding.EncodeToString([]byte(b.Username+":"+b.Password))
	return map[string]string{"authorization": cred}, nil
}

func (b *BasicAuth) RequireTransportSecurity() bool {
	return false // 若启用 TLS，建议返回 true
}
