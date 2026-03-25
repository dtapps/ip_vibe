package grpc_interceptor

import (
	"context"
	"log/slog"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log/operationlog"

	"google.golang.org/grpc"
)

// 审计上下文拦截器
type AuditContextInterceptor struct {
	logger *slog.Logger
	source operationlog.Source
}

// NewAuditContextInterceptor 创建审计上下文拦截器
func NewAuditContextInterceptor(logger *slog.Logger, source operationlog.Source) *AuditContextInterceptor {
	if source == operationlog.SourceUnknown {
		source = operationlog.SourceGrpc
	}
	return &AuditContextInterceptor{
		logger: logger,
		source: source,
	}
}

// 一元调用
func (aci *AuditContextInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

		// 设置来源设为 Grpc
		newCtx := schemahook.WithSource(ctx, aci.source.String())

		// 调用处理
		return handler(newCtx, req)
	}
}

// auditContextServerStream 包装 grpc.ServerStream，重写 Context()
type auditContextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *auditContextServerStream) Context() context.Context {
	return s.ctx
}

// 流式调用
func (aci *AuditContextInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 将 source 注入到 context 中
		newCtx := schemahook.WithSource(stream.Context(), aci.source.String())

		// 创建一个包装后的 ServerStream，其 Context() 返回 newCtx
		wrappedStream := &auditContextServerStream{
			ServerStream: stream,
			ctx:          newCtx,
		}

		// 调用处理
		return handler(srv, wrappedStream)
	}
}
