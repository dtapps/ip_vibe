package grpc_interceptor

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggerUnaryInterceptor 日志 一元拦截器
func LoggerUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		// 从 FullMethod 解析：/package.service/method
		full := info.FullMethod               // e.g. "/helloworld.Greeter/SayHello"
		parts := strings.SplitN(full, "/", 3) // ["", "helloworld.Greeter", "SayHello"]
		svc := "-"
		mtd := "-"
		if len(parts) == 3 {
			svc, mtd = parts[1], parts[2]
		}

		st, _ := status.FromError(err)
		code := st.Code()

		// 记录日志
		if err != nil && code != codes.OK {
			logger.ErrorContext(ctx, "gRPC 调用错误",
				slog.Duration("duration", duration),
				slog.String("service", svc),
				slog.String("method", mtd),
				slog.String("code", code.String()),
				slog.String("address", peerAddr(ctx)),
				slog.Any("err", err),
			)
		} else {
			logger.DebugContext(ctx, "gRPC 调用完成",
				slog.Duration("duration", duration),
				slog.String("service", svc),
				slog.String("method", mtd),
				slog.String("code", code.String()),
				slog.String("address", peerAddr(ctx)),
			)
		}

		return resp, err
	}
}

// peerAddr 从 ctx 解析对端地址（简化示例）
func peerAddr(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		return p.Addr.String()
	}
	return "-"
}
