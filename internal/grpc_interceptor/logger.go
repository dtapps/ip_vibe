package grpc_interceptor

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// 日志拦截器
type LoggingInterceptor struct {
	logger *slog.Logger
}

// NewLoggingInterceptor 创建日志拦截器
func NewLoggingInterceptor(logger *slog.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

// 一元调用
func (li *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 获取对端地址
		address := li.peerAddr(ctx)

		// 构建日志字段
		loggerFields := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.String("address", address),
		}

		// 开始日志
		li.logger.LogAttrs(ctx, slog.LevelInfo, "开始 gRPC 一元调用", loggerFields...)

		start := time.Now()

		// 调用处理
		resp, err := handler(ctx, req)

		duration := time.Since(start)

		loggerFields = append(loggerFields, slog.Duration("duration", duration))

		st, _ := status.FromError(err)
		code := st.Code()

		loggerFields = append(loggerFields, slog.String("code", code.String()))

		// 结束日志
		if err != nil && code != codes.OK {
			loggerFields = append(loggerFields, slog.Any("err", err))
			li.logger.LogAttrs(ctx, slog.LevelInfo, "异常结束 gRPC 一元调用", loggerFields...)
		} else {
			li.logger.LogAttrs(ctx, slog.LevelInfo, "结束 gRPC 一元调用", loggerFields...)
		}

		return resp, err
	}
}

// 流式调用
func (li *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		// 获取上下文
		ctx := ss.Context()

		// 获取对端地址
		address := li.peerAddr(ctx)

		// 构建日志字段
		loggerFields := []slog.Attr{
			slog.String("method", info.FullMethod),
			slog.String("address", address),
		}

		// 开始日志
		li.logger.LogAttrs(ctx, slog.LevelInfo, "开始 gRPC 流式调用", loggerFields...)

		start := time.Now()

		// 调用处理
		err := handler(srv, ss)

		duration := time.Since(start)

		loggerFields = append(loggerFields, slog.Duration("duration", duration))

		st, _ := status.FromError(err)
		code := st.Code()

		loggerFields = append(loggerFields, slog.String("code", code.String()))

		// 结束日志
		if err != nil && code != codes.OK {
			loggerFields = append(loggerFields, slog.Any("err", err))
			li.logger.LogAttrs(ctx, slog.LevelError, "异常结束 gRPC 流式调用", loggerFields...)
		} else {
			li.logger.LogAttrs(ctx, slog.LevelInfo, "结束 gRPC 流式调用", loggerFields...)
		}

		return err
	}
}

// peerAddr 从 ctx 解析对端地址
func (li *LoggingInterceptor) peerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "-"
}
