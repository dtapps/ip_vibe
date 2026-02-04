package manager

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"vibe_ip/internal/gin_server"
	"vibe_ip/internal/grpc_server"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type MixedServerManager struct {
	ginMgr  *gin_server.GinServerManager
	grpcMgr *grpc_server.GRPCServerManager
	server  *http.Server
	logger  *slog.Logger
	addr    string
}

func NewMixedServerManager(addr string, ginMgr *gin_server.GinServerManager, grpcMgr *grpc_server.GRPCServerManager, logger *slog.Logger) *MixedServerManager {
	return &MixedServerManager{
		addr:    addr,
		ginMgr:  ginMgr,
		grpcMgr: grpcMgr,
		logger:  logger,
	}
}

// Run 同时启动 Gin 和 gRPC 在同一个端口
func (m *MixedServerManager) Run(ctx context.Context) {
	// 1. 获取两个实例
	ginEngine := m.ginMgr.GetServer()
	grpcServer := m.grpcMgr.GetServer()

	// 2. 构造分流处理器
	mixedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// gRPC 特征判断：HTTP/2 协议且 Content-Type 是 application/grpc
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			ginEngine.ServeHTTP(w, r)
		}
	})

	// 3. 配置支持 h2c 的 Server
	// h2c.NewHandler 会自动根据请求升降级协议
	m.server = &http.Server{
		Addr:    m.addr,
		Handler: h2c.NewHandler(mixedHandler, &http2.Server{}),
	}

	m.logger.InfoContext(ctx, "混合服务启动成功 (Gin + gRPC 共享端口)",
		slog.String("addr", m.addr),
	)

	// 4. 启动监听
	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		m.logger.ErrorContext(ctx, "混合服务运行失败", slog.Any("err", err))
		panic(err)
	}
}

// Stop 优雅关闭所有关联服务
func (m *MixedServerManager) Stop(ctx context.Context) {
	if m.server != nil {
		m.logger.InfoContext(ctx, "混合服务正在优雅关闭...")

		// 关闭外层 HTTP Server
		if err := m.server.Shutdown(ctx); err != nil {
			m.logger.ErrorContext(ctx, "混合服务关闭失败", slog.Any("err", err))
		}
	}

	// 触发子管理器的清理逻辑
	// m.grpcMgr.Stop(ctx)
	// m.ginMgr.Stop(ctx)

	m.logger.InfoContext(ctx, "混合服务已安全停止")
}
