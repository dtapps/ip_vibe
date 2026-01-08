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
	Addr     string // 地址
	Username string // 用户名
	Password string // 密码
}

// GRPCServerManager gRPC 服务管理器
type GRPCServerManager struct {
	server   *grpc.Server // gRPC 服务实例
	listener net.Listener // 监听套接字
	logger   *slog.Logger // 日志记录器
	config   Config       // gRPC 服务配置参数
}

var (
	instance *GRPCServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *GRPCServerManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &GRPCServerManager{
			logger: logger,
			config: cfg,
		}
	})
	return instance
}

// GetServer 获取底层的 gRPC 服务实例
func (m *GRPCServerManager) GetServer() *grpc.Server {
	if m.server == nil {
		return nil
	}
	return m.server
}

// Run 启动 gRPC 服务
func (m *GRPCServerManager) Run(ctx context.Context) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "gRPC 服务不启动，未配置地址")
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
		logger.ErrorContext(ctx, "gRPC 服务启动失败",
			slog.Any("err", err),
		)
		panic(err)
	}
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
