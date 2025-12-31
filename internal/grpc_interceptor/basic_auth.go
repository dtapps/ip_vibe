package grpc_interceptor

import (
	"context"
	"encoding/base64"
	"log/slog"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// BasicAuth 返回一个 gRPC 拦截器：仅允许通过 Basic 认证的用户访问指定路径
func BasicAuth(logger *slog.Logger, username string, password string) grpc.UnaryServerInterceptor {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 如果用户名或密码为空就不校验
		if username == "" || password == "" {
			return handler(ctx, req)
		}

		// 提取 Authorization 头（basic scheme）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少或无效的 Authorization 头")
		}
		vals := md["authorization"]
		if len(vals) == 0 {
			return nil, status.Error(codes.Unauthenticated, "缺少或无效的 Authorization 头")
		}
		authHeader := vals[0]
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			return nil, status.Error(codes.Unauthenticated, "缺少或无效的 Authorization 头")
		}

		// Base64 解码
		decoded, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "凭证解码失败")
		}

		// 分离 username:password
		parts := strings.SplitN(string(decoded), ":", 2)
		if len(parts) != 2 {
			return nil, status.Error(codes.Unauthenticated, "无效凭证格式")
		}

		// 校验
		user, pass := parts[0], parts[1]
		if user != username || pass != password {
			return nil, status.Error(codes.Unauthenticated, "用户名或密码错误")
		}

		return handler(ctx, req)
	}
}
