package grpc_client

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/grpc_credential"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

// Config gRPC 客户端所需的配置参数
type Config struct {
	Addr     string // 地址
	Username string // 用户名
	Password string // 密码
}

// GRPCClientManager gRPC 客户端管理器
type GRPCClientManager struct {
	client *grpc.ClientConn // gRPC 客户端连接
	logger *slog.Logger     // 日志记录器
	config Config           // gRPC 客户端配置参数
}

var (
	instance *GRPCClientManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *GRPCClientManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		instance = &GRPCClientManager{
			logger: logger,
			config: cfg,
		}
	})
	return instance
}

// Run 启动 gRPC 客户端
func (m *GRPCClientManager) Run(ctx context.Context) (err error) {

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "gRPC 客户端不启动，未配置地址")
		return nil
	}

	// 配置选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		// 鉴权
		grpc.WithPerRPCCredentials(grpc_credential.NewClientBasicAuth(
			m.logger,
			m.config.Username,
			m.config.Password,
		)),
		// 重试
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
	}

	// 连接 gRPC 服务器
	m.client, err = grpc.NewClient(m.config.Addr, opts...)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC 客户端连接服务器失败",
			slog.Any("err", err),
		)
		return err
	}

	return nil
}

// GetServer 获取底层的 grpc.Server 实例
// 注意：如果在 Run 之前调用，可能会返回 nil
func (m *GRPCClientManager) GetClient() *grpc.ClientConn {
	return m.client
}

// Stop 优雅关闭
func (m *GRPCClientManager) Stop(ctx context.Context) {
	if m.client != nil {
		m.logger.InfoContext(ctx, "gRPC 客户端正在优雅关闭...")
		m.client.Close()
	}
	m.logger.InfoContext(ctx, "gRPC 客户端已安全停止")
}
