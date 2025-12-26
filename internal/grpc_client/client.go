package grpc_client

import (
	"fmt"

	"google.golang.org/grpc"
)

// GRPCClient gRPC 客户端
type GRPCClient struct {
	client *grpc.ClientConn
	opts   []grpc.DialOption
}

// NewGRPCClient 创建 gRPC 客户端
func NewGRPCClient() *GRPCClient {
	return &GRPCClient{}
}

// AddDialOption 添加 gRPC 连接选项
func (c *GRPCClient) AddDialOption(opts ...grpc.DialOption) {
	c.opts = append(c.opts, opts...)
}

// Close 关闭 gRPC 客户端
func (c *GRPCClient) Close() error {
	return c.client.Close()
}

// Dial 连接 gRPC 服务器
func (c *GRPCClient) Dial(addr string) error {
	client, err := grpc.NewClient(addr, c.opts...)
	if err != nil {
		return fmt.Errorf("连接 gRPC 服务器失败: %w", err)
	}
	c.client = client
	return nil
}

// GetClient 获取 gRPC 客户端
func (c *GRPCClient) GetClient() *grpc.ClientConn {
	return c.client
}
