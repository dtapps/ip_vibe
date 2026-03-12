package mixed_manager

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
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

var (
	instanceMixedServer *MixedServerManager
	onceMixedServer     sync.Once
)

// NewMixedServerManager 创建混合服务管理器实例
func NewMixedServerManager(addr string, ginMgr *gin_server.GinServerManager, grpcMgr *grpc_server.GRPCServerManager, logger *slog.Logger) *MixedServerManager {
	onceMixedServer.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instanceMixedServer = &MixedServerManager{
			addr:    addr,
			ginMgr:  ginMgr,
			grpcMgr: grpcMgr,
			logger:  logger,
		}
	})
	return instanceMixedServer
}

// GetMixedServerInstance 获取混合服务管理器单例
func GetMixedServerInstance() *MixedServerManager {
	return instanceMixedServer
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
		Addr: m.addr,
		Handler: h2c.NewHandler(mixedHandler, &http2.Server{
			// 保活核心：30秒无数据则发PING探测
			ReadIdleTimeout: 30 * time.Second,
			// PING响应超时：15秒没回应就断开（清理僵尸连接）
			PingTimeout: 15 * time.Second,
			// 可选：限制并发流
			MaxConcurrentStreams: 100,
		}),

		// 应用层超时（必须配置）
		// ReadTimeout:  30 * time.Second,
		// WriteTimeout: 30 * time.Second,
		IdleTimeout: 120 * time.Second, // 空闲连接保持时间（≥ Nginx proxy_read_timeout）

		// 可选：连接状态日志
		ConnState: func(conn net.Conn, state http.ConnState) {
			switch state {
			case http.StateClosed:
				m.logger.DebugContext(ctx, "连接正常关闭",
					slog.String("remote", conn.RemoteAddr().String()),
				)
			case http.StateHijacked:
				m.logger.WarnContext(ctx, "连接被劫持（如 WebSocket）",
					slog.String("remote", conn.RemoteAddr().String()),
				)
			}
		},
	}

	m.logger.InfoContext(ctx, "混合服务启动成功 (Gin + MCP + gRPC 共享端口)",
		slog.String("addr", m.addr),
	)

	// 4. 启动监听
	err := m.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		m.logger.ErrorContext(ctx, "混合服务运行失败",
			slog.Any("err", err),
		)
		panic(err)
	}
}

// Stop 优雅关闭所有关联服务
func (m *MixedServerManager) Stop(ctx context.Context) {
	if m.server != nil {
		m.logger.InfoContext(ctx, "混合服务正在优雅关闭...")

		// 关闭外层 HTTP Server
		err := m.server.Shutdown(ctx)
		if err != nil {
			m.logger.ErrorContext(ctx, "混合服务关闭失败",
				slog.Any("err", err),
			)
		}
	}

	m.logger.InfoContext(ctx, "混合服务已安全停止")
}
