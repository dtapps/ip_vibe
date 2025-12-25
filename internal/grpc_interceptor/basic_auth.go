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
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 如果用户名或密码为空就不校验
		if username == "" || password == "" {
			return handler(ctx, req)
		}

		// 读取 Authorization 头
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("missing metadata", slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}
		vals := md["authorization"]
		if len(vals) == 0 {
			logger.Warn("missing authorization header", slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}
		authHeader := vals[0]

		// 解析 Basic
		const prefix = "Basic "
		if !strings.HasPrefix(authHeader, prefix) {
			logger.Warn("invalid authorization scheme", slog.String("auth", authHeader), slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid authorization scheme")
		}
		b, err := base64.StdEncoding.DecodeString(authHeader[len(prefix):])
		if err != nil {
			logger.Warn("base64 decode failed", slog.String("err", err.Error()), slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid base64")
		}
		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			logger.Warn("invalid basic auth format", slog.String("auth", authHeader), slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid basic auth format")
		}
		user, pass := pair[0], pair[1]

		// 校验
		if user != username || pass != password {
			logger.Warn("invalid username or password", slog.String("user", user), slog.String("method", info.FullMethod))
			return nil, status.Error(codes.Unauthenticated, "invalid username or password")
		}

		logger.Info("basic auth passed", slog.String("user", user), slog.String("method", info.FullMethod))
		return handler(ctx, req)
	}
}
