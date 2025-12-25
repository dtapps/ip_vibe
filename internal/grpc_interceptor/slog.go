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

// UnaryLogger 返回一个 unary 拦截器：记录开始、结束、耗时与状态码
func UnaryLogger(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(start).Milliseconds()

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

		attrs := []slog.Attr{
			slog.String("grpc.service", svc),
			slog.String("grpc.method", mtd),
			slog.String("grpc.code", code.String()),
			slog.Int64("grpc.time_ms", dur),
			slog.String("peer.address", peerAddr(ctx)),
		}
		if err != nil && code != codes.OK {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		logger.LogAttrs(ctx, slog.LevelInfo, "finished unary call", attrs...)
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
