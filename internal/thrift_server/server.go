package thrift_server

import (
	"context"
	"log/slog"
	"sync"
	"vibe_ip/internal/thrift_generated/open"
	"vibe_ip/internal/thrift_process"
	"vibe_ip/internal/thrift_service"

	"github.com/apache/thrift/lib/go/thrift"
)

// Config Thrift 服务所需的配置参数
type Config struct {
	Addr string
}

// ThriftServerManager Thrift 服务管理器
type ThriftServerManager struct {
	server *thrift.TSimpleServer // Thrift 服务器实例
	logger *slog.Logger          // 日志记录器
	config Config                // Thrift 服务配置参数
}

var (
	instance *ThriftServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *ThriftServerManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &ThriftServerManager{
			logger: logger,
			config: cfg,
		}
	})
	return instance
}

// Run 启动 Thrift 服务
func (m *ThriftServerManager) Run(ctx context.Context) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "Thrift 服务不启动，未配置地址")
		return
	}

	// 启动 Thrift 服务
	socket, err := thrift.NewTServerSocket(m.config.Addr)
	if err != nil {
		logger.ErrorContext(ctx, "Thrift 服务监听失败",
			slog.Any("err", err),
		)
		return
	}

	// 传输工厂
	transportFactory := thrift.NewTFramedTransportFactoryConf(
		thrift.NewTTransportFactory(),
		&thrift.TConfiguration{},
	)

	// 协议工厂
	protocolFactory := thrift.NewTBinaryProtocolFactoryConf(&thrift.TConfiguration{})

	// 业务逻辑层 (Handler)
	openHandler := thrift_service.NewOpenServer(m.logger)
	processor := open.NewOpenServiceProcessor(openHandler)

	// 挂载【通用】处理器包装器
	// 无论 IDL 怎么变，这里一行代码搞定所有服务的日志和耗时
	wrappedProcessor := thrift_process.NewGenericInterceptor(processor, m.logger)

	// 传入 SimpleServer
	m.server = thrift.NewTSimpleServer4(
		wrappedProcessor,
		socket,
		transportFactory,
		protocolFactory,
	)

	logger.InfoContext(ctx, "Thrift 服务启动成功")

	if err := m.server.Serve(); err != nil {
		logger.ErrorContext(ctx, "Thrift 服务运行异常",
			slog.Any("err", err),
		)
	}
}

// GetServer 获取底层的 thrift.TSimpleServer 实例
// 注意：如果在 Run 之前调用，可能会返回 nil
func (m *ThriftServerManager) GetServer() *thrift.TSimpleServer {
	return m.server
}

// Stop 优雅关闭
func (m *ThriftServerManager) Stop(ctx context.Context) {
	if m.server != nil {
		m.logger.InfoContext(ctx, "Thrift 服务正在优雅关闭...")
		m.server.Stop()
	}
	m.logger.InfoContext(ctx, "Thrift 服务已安全停止")
}
