package grpc_credential

import (
	"context"
	"log/slog"
	"testing"

	"vibe_ip/internal/core"
)

func TestClientBasicAuthGetRequestMetadata(t *testing.T) {
	logger := slog.New(slog.DiscardHandler)

	tests := []struct {
		name        string
		username    string
		password    string
		wantNilMeta bool
	}{
		{name: "用户名密码齐全注入 authorization", username: "admin", password: "123456"},
		{name: "用户名为空不注入", username: "", password: "123456", wantNilMeta: true},
		{name: "密码为空不注入", username: "admin", password: "", wantNilMeta: true},
		{name: "两者都为空不注入", username: "", password: "", wantNilMeta: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cred := NewClientBasicAuth(logger, tt.username, tt.password)
			md, err := cred.GetRequestMetadata(context.Background(), "/vibe_ip.Service/Method")
			if err != nil {
				t.Fatalf("GetRequestMetadata 意外报错: %v", err)
			}
			if tt.wantNilMeta {
				if md != nil {
					t.Fatalf("期望返回 nil metadata，实际为 %v", md)
				}
				return
			}

			want := core.GetBasicAuthHeader(core.GetBasicAuthToken(tt.username, tt.password))
			if got := md["authorization"]; got != want {
				t.Fatalf("authorization = %q, want %q", got, want)
			}
			if len(md) != 1 {
				t.Fatalf("metadata 应只有 authorization 一项，实际为 %v", md)
			}
		})
	}
}

func TestClientBasicAuthRequireTransportSecurity(t *testing.T) {
	cred := NewClientBasicAuth(slog.New(slog.DiscardHandler), "admin", "123456")
	if cred.RequireTransportSecurity() {
		t.Fatal("RequireTransportSecurity 应为 false")
	}
}
