package thrift_client

import (
	"context"
	"log/slog"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
)

// Config Thrift 客户端所需的配置参数
type Config struct {
	Addr     string // 地址
	Username string // 用户名
	Password string // 密码
}

// ThriftClientManager Thrift 客户端管理器
type ThriftClientManager struct {
	client *thrift.TStandardClient // Thrift 客户端连接
	logger *slog.Logger            // 日志记录器
	config Config                  // Thrift 客户端配置参数
}

var (
	instance *ThriftClientManager
	once     sync.Once
)

// NewManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *ThriftClientManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &ThriftClientManager{
			logger: logger,
			config: cfg,
		}
	})
	return instance
}

// Conn 连接 Thrift 客户端
func (m *ThriftClientManager) Conn(ctx context.Context) (err error) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "Thrift 客户端不连接，未配置地址")
		return nil
	}

	// 设置传输层
	conf := &thrift.TConfiguration{}
	socket := thrift.NewTSocketConf(m.config.Addr, conf)

	// 打开连接
	transport := thrift.NewTFramedTransportConf(socket, conf)
	if err := transport.Open(); err != nil {
		logger.ErrorContext(ctx, "Thrift 客户端连接服务器失败",
			slog.Any("err", err),
		)
		return err
	}

	// 设置协议层
	protocol := thrift.NewTBinaryProtocolConf(transport, conf)

	// 创建客户端
	m.client = thrift.NewTStandardClient(protocol, protocol)

	return nil
}

// GetClient 获取底层的 Thrift 客户端连接
func (m *ThriftClientManager) GetClient() *thrift.TStandardClient {
	if m.client == nil {
		return nil
	}
	return m.client
}

// Stop 优雅关闭
func (m *ThriftClientManager) Stop(ctx context.Context) {
	if m.client != nil {
		m.logger.InfoContext(ctx, "Thrift 客户端正在优雅关闭...")
	}
	m.logger.InfoContext(ctx, "Thrift 客户端管理器已停止")
}
