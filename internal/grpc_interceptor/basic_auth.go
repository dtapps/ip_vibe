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

// BasicAuth拦截器
type BasicAuthInterceptor struct {
	logger   *slog.Logger
	authKey  string
	username string
	password string
}

// NewBasicAuthInterceptor 创建BasicAuth拦截器
func NewBasicAuthInterceptor(logger *slog.Logger, username string, password string) *BasicAuthInterceptor {
	return &BasicAuthInterceptor{
		logger:   logger,
		authKey:  "authorization",
		username: username,
		password: password,
	}
}

// 一元调用
func (bai *BasicAuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 如果用户名或密码为空就不校验
		if bai.username == "" || bai.password == "" {
			return handler(ctx, req)
		}

		// 提取 Authorization 头（basic scheme）
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			err := fmt.Errorf("缺少或无效的 Authorization 头")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		vals := md[bai.authKey]
		if len(vals) == 0 {
			err := fmt.Errorf("缺少或无效的 Authorization 头")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// 解析 Basic Auth
		user, pass, err := core.ParseBasicAuthToken(vals[0])
		if err != nil {
			bai.logger.ErrorContext(ctx, "[BasicAuth] 解析 Authorization 头失败",
				slog.String("err", err.Error()),
			)
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		if user != bai.username || pass != bai.password {
			err := fmt.Errorf("用户名或密码错误")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 用户名或密码错误")
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// 认证通过，继续处理流
		return handler(ctx, req)
	}
}

// 流式调用
func (bai *BasicAuthInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		// 如果用户名或密码为空就不校验
		if bai.username == "" || bai.password == "" {
			return handler(srv, ss)
		}

		// 从流的上下文中提取 metadata
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			err := fmt.Errorf("缺少或无效的 Authorization 头")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return status.Error(codes.Unauthenticated, err.Error())
		}

		vals := md[bai.authKey]
		if len(vals) == 0 {
			err := fmt.Errorf("缺少或无效的 Authorization 头")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 缺少或无效的 Authorization 头")
			return status.Error(codes.Unauthenticated, err.Error())
		}

		// 解析 Basic Auth
		user, pass, err := core.ParseBasicAuthToken(vals[0])
		if err != nil {
			bai.logger.ErrorContext(ctx, "[BasicAuth] 解析 Authorization 头失败",
				slog.String("err", err.Error()),
			)
			return status.Error(codes.Unauthenticated, err.Error())
		}

		if user != bai.username || pass != bai.password {
			err := fmt.Errorf("用户名或密码错误")
			bai.logger.ErrorContext(ctx, "[BasicAuth] 用户名或密码错误")
			return status.Error(codes.Unauthenticated, err.Error())
		}

		// 认证通过，继续处理流
		return handler(srv, ss)
	}
}
