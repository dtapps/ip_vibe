package grpc_interceptor

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/core"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// BasicAuthUnaryInterceptor BasicAuth 一元拦截器
func BasicAuthUnaryInterceptor(logger *slog.Logger, username string, password string) grpc.UnaryServerInterceptor {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 认证头键名
	const authKey = "authorization"

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 如果用户名或密码为空就不校验
		if username == "" || password == "" {
			return handler(ctx, req)
		}

		// 提取 Authorization 头（basic scheme）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			err := fmt.Errorf("缺少或无效的 Authorization 头")

			logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		vals := md[authKey]
		if len(vals) == 0 {
			err := fmt.Errorf("缺少或无效的 Authorization 头")

			logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// 解析
		user, pass, err := core.ParseBasicAuthToken(vals[0])
		if err != nil {
			logger.ErrorContext(ctx, "[BasicAuth] 解析 Authorization 头失败",
				slog.String("err", err.Error()),
			)
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if user != username || pass != password {
			err := fmt.Errorf("用户名或密码错误")

			logger.ErrorContext(ctx, "[BasicAuth] 用户名或密码错误")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		return handler(ctx, req)
	}
}
