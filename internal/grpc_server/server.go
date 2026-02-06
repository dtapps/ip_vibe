package grpc_server

import (
	"context"
	"log/slog"
	"net"
	"sync"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
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
	server   *grpc.Server    // gRPC 服务实例
	listener net.Listener    // 监听套接字
	db       *ent.Client     // 数据库客户端
	logDb    *ent_log.Client // 日志数据库客户端
	logger   *slog.Logger    // 日志记录器
	config   Config          // gRPC 服务配置参数

	deviceService    *grpc_service.DeviceService
	openService      *grpc_service.OpenService
	remoteLogService *grpc_service.RemoteLogService
}

var (
	instance *GRPCServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(db *ent.Client, logDb *ent_log.Client, logger *slog.Logger, cfg Config) *GRPCServerManager {
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

// GetGRPCServer 获取底层的 gRPC 服务实例
func (m *GRPCServerManager) GetGRPCServer() *grpc.Server {
	if m.server == nil {
		return nil
	}
	return m.server
}

// initServer 初始化 gRPC 服务
func (m *GRPCServerManager) initServer() {
	// 审计上下文拦截器
	auditInterceptor := grpc_interceptor.NewAuditContextInterceptor(
		m.logger,
		operationlog.SourceGrpc,
	)

	// 日志拦截器
	loggingInterceptor := grpc_interceptor.NewLoggingInterceptor(m.logger)

	// BasicAuth 拦截器
	basicAuthInterceptor := grpc_interceptor.NewBasicAuthInterceptor(
		m.logger,
		m.config.Username,
		m.config.Password,
	)

	// 配置一元拦截器
	usis := []grpc.UnaryServerInterceptor{
		// 审计上下文
		auditInterceptor.Unary(),
		// 日志
		loggingInterceptor.Unary(),
		// BasicAuth
		basicAuthInterceptor.Unary(),
	}

	// 配置流式拦截器
	ssis := []grpc.StreamServerInterceptor{
		// 审计上下文
		auditInterceptor.Stream(),
		// 日志
		loggingInterceptor.Stream(),
		// BasicAuth
		basicAuthInterceptor.Stream(),
	}

	// 配置选项
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(usis...),
		grpc.ChainStreamInterceptor(ssis...),
		grpc.MaxRecvMsgSize(10 * 1024 * 1024),
		grpc.MaxSendMsgSize(10 * 1024 * 1024),
	}

	m.server = grpc.NewServer(opts...)

	// 注册服务
	m.deviceService = grpc_service.NewDeviceService(m.logger)
	m.openService = grpc_service.NewOpenService(m.db, m.logDb, m.logger)
	m.remoteLogService = grpc_service.NewRemoteLogService(m.logger)
	m.deviceService.RegisterService(m.server)
	m.openService.RegisterService(m.server)
	m.remoteLogService.RegisterService(m.server)
}

// GetServer 获取 gRPC 服务，方便其他服务使用
func (m *GRPCServerManager) GetServer() *grpc.Server {
	if m.server == nil {
		m.initServer()
	}
	return m.server
}

// Run 启动 gRPC 服务
func (m *GRPCServerManager) Run(ctx context.Context) {

	// 确保 server 已初始化
	m.GetServer()

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "gRPC 服务启动错误，未配置地址")
		return
	}

	var err error
	m.listener, err = net.Listen("tcp", m.config.Addr)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC 服务监听失败",
			slog.Any("err", err),
		)
		panic(err)
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
		if err := m.listener.Close(); err != nil {
			m.logger.ErrorContext(ctx, "gRPC 服务关闭监听失败",
				slog.Any("err", err),
			)
		}
	}
	m.logger.InfoContext(ctx, "gRPC 服务已安全停止")
}

// GetDeviceService 返回 DeviceService 实例
func (m *GRPCServerManager) GetDeviceService() *grpc_service.DeviceService {
	return m.deviceService
}

// GetOpenService 返回 OpenService 实例
func (m *GRPCServerManager) GetOpenService() *grpc_service.OpenService {
	return m.openService
}

// GetRemoteLogService 返回 RemoteLogService 实例
func (m *GRPCServerManager) GetRemoteLogService() *grpc_service.RemoteLogService {
	return m.remoteLogService
}
