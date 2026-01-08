package grpc_client

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/grpc_credential"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type BasicAuth struct {
	Username string // 用户名
	Password string // 密码
}

type Options struct {
	Addr      string       // 服务器地址
	BasicAuth BasicAuth    // 基本认证信息
	Logger    *slog.Logger // 日志记录器
}

type Client struct {
	client *grpc.ClientConn
	opts   []grpc.DialOption
}

// Run 运行 gRPC 客户端
func Run(ctx context.Context, opts Options) (*Client, error) {

	// 检查服务器地址
	if opts.Addr == "" {
		return nil, fmt.Errorf("服务器地址不能为空")
	}

	// 检查日志记录器
	if opts.Logger == nil {
		opts.Logger = slog.Default()
	}

	c := &Client{}

	c.opts = append(
		c.opts,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	// 鉴权
	if opts.BasicAuth.Username != "" && opts.BasicAuth.Password != "" {
		c.opts = append(
			c.opts,
			grpc.WithPerRPCCredentials(grpc_credential.NewClientBasicAuth(
				opts.Logger,
				opts.BasicAuth.Username,
				opts.BasicAuth.Password,
			)),
		)
	}

	// 重试
	c.opts = append(
		c.opts,
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
	)
	c.opts = append(
		c.opts,
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
	)

	// 连接 gRPC 服务器
	client, err := grpc.NewClient(opts.Addr, c.opts...)
	if err != nil {
		return nil, fmt.Errorf("连接服务器失败: %w", err)
	}
	c.client = client

	return c, nil
}

// Close 关闭 gRPC 客户端
func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}

// GetClient 获取 gRPC 客户端
func (c *Client) GetClient() *grpc.ClientConn {
	if c.client == nil {
		return nil
	}
	return c.client
}
