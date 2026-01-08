package grpc_service

import (
	"context"
	"log/slog"
	"vibe_ip/internal/grpc_generated"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// IndexServer 实现 gRPC 服务接口
type IndexServer struct {
	grpc_generated.UnimplementedIndexServiceServer              // 嵌入未实现的服务
	logger                                         *slog.Logger // 日志记录器
}

// NewIndexServer 创建 gRPC 服务实例
func NewIndexServer(logger *slog.Logger) *IndexServer {
	return &IndexServer{
		logger: logger,
	}
}

// RegisterService 注册服务到 gRPC
func (s *IndexServer) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterIndexServiceServer(server, s)
}

// Index 实现 入口 方法
func (s *IndexServer) Index(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	// 转换为 gRPC 响应
	return &emptypb.Empty{}, nil
}
