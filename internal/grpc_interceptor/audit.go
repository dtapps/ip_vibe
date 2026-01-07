package grpc_interceptor

import (
	"context"
	"log/slog"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log/operationlog"

	"google.golang.org/grpc"
)

// AuditContextInterceptor 审计上下文 一元拦截器
func AuditContextInterceptor(logger *slog.Logger, source operationlog.Source) grpc.UnaryServerInterceptor {

	if source == operationlog.SourceUnknown {
		source = operationlog.SourceGrpc
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 设置来源设为 Grpc
		newCtx := schemahook.WithSource(ctx, operationlog.SourceGrpc.String())

		// 使用注入了标识的 newCtx 继续执行
		return handler(newCtx, req)
	}
}
