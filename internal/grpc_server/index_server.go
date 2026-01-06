package grpc_server

import (
	"context"
	"vibe_ip/internal/grpc_generated"

	"github.com/dubbogo/grpc-go/codes"
	"google.golang.org/grpc"
)

// IndexServer 实现 gRPC 服务接口
type IndexServer struct {
	grpc_generated.UnimplementedIndexServiceServer
}

// NewIndexServer 创建 gRPC 服务实例
func NewIndexServer() *IndexServer {
	return &IndexServer{}
}

// RegisterService 注册 gRPC 服务到服务器
func (s *IndexServer) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterIndexServiceServer(server, s)
}

// Index 实现 入口 方法
func (s *IndexServer) Index(ctx context.Context, req *grpc_generated.IndexRequest) (*grpc_generated.IndexResponse, error) {
	// 转换为 gRPC 响应
	return &grpc_generated.IndexResponse{
		Code:    uint32(codes.OK),
		Message: "grpc ok",
	}, nil
}
