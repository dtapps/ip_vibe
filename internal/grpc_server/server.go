package grpc_server

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/grpc_interceptor"
	"vibe_ip/internal/grpc_service"

	"google.golang.org/grpc"
)

// Config gRPC 服务所需的配置参数
type Config struct {
	Addr     string
	Username string
	Password string
}

// GRPCServerManager gRPC 服务管理器
type GRPCServerManager struct {
	server   *grpc.Server
	listener net.Listener
	logger   *slog.Logger
	config   Config
}

var (
	instance *GRPCServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *GRPCServerManager {
	once.Do(func() {
		instance = &GRPCServerManager{
			logger: logger,
			config: cfg,
		}
	})
	return instance
}

// Run 启动服务
func (m *GRPCServerManager) Run(ctx context.Context) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	if m.config.Addr == "" {
		logger.InfoContext(ctx, "gRPC 服务不启动，未配置地址")
		return
	}

	// 配置拦截器
	usis := []grpc.UnaryServerInterceptor{
		// 审计上下文
		grpc_interceptor.AuditContextInterceptor(m.logger, operationlog.SourceGrpc),
		// BasicAuth
		grpc_interceptor.BasicAuthUnaryInterceptor(
			m.logger,
			m.config.Username,
			m.config.Password,
		),
		// 日志
		grpc_interceptor.LoggerUnaryInterceptor(m.logger),
	}

	// 配置选项
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(usis...),
		grpc.MaxRecvMsgSize(10 * 1024 * 1024),
		grpc.MaxSendMsgSize(10 * 1024 * 1024),
	}

	m.server = grpc.NewServer(opts...)

	// 注册服务
	grpc_service.NewIndexServer().RegisterService(m.server)
	grpc_service.NewOpenServer().RegisterService(m.server)

	var err error
	m.listener, err = net.Listen("tcp", m.config.Addr)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC 服务监听失败",
			slog.Any("err", err),
		)
		return
	}

	logger.InfoContext(ctx, "gRPC 服务启动成功")

	if err := m.server.Serve(m.listener); err != nil {
		logger.ErrorContext(ctx, "gRPC 服务运行异常",
			slog.Any("err", err),
		)
	}
}

// GetServer 获取底层的 grpc.Server 实例
// 注意：如果在 Run 之前调用，可能会返回 nil
func (m *GRPCServerManager) GetServer() *grpc.Server {
	return m.server
}

// Stop 优雅关闭
func (m *GRPCServerManager) Stop(ctx context.Context) {
	if m.server != nil {
		m.logger.InfoContext(ctx, "gRPC 服务正在优雅关闭...")
		m.server.GracefulStop()
	}
	if m.listener != nil {
		m.listener.Close()
	}
	m.logger.InfoContext(ctx, "gRPC 服务已安全停止")
}
