package grpc_server

import (
	"context"
	"vibe_ip/internal/backend/service"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/types"

	"google.golang.org/grpc"
)

// OpenServer 实现 gRPC 服务
type OpenServer struct {
	grpc_generated.UnimplementedOpenServiceServer
	deviceService service.DeviceOpenService
}

// NewOpenServer 创建新的 gRPC 服务实例
func NewOpenServer() *OpenServer {
	// 使用与你的 HTTP 处理器相同的服务实例
	deviceService := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
		webhookService.NewSend(global.GvaEntDb, global.GvaEntLogDb),
	)

	return &OpenServer{
		deviceService: deviceService,
	}
}

// RegisterService 注册 gRPC 服务到服务器
func (s *OpenServer) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterOpenServiceServer(server, s)
}

// ReportDeviceIP 实现设备上报IP的 gRPC 方法
func (s *OpenServer) ReportDeviceIP(ctx context.Context, req *grpc_generated.ReportDeviceIPRequest) (*grpc_generated.ReportDeviceIPResponse, error) {
	// 将 gRPC 请求转换为你的 types 结构
	request := &types.DeviceReportDeviceIpRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		DeviceName: req.DeviceName,
		IPV4Addr:   req.Ipv4Addr,
		IPV6Addr:   req.Ipv6Addr,
	}

	// 调用你现有的业务逻辑（与 HTTP 处理器使用相同的代码）
	response, err := s.deviceService.DeviceReportDeviceIp(ctx, request)
	if err != nil {
		return nil, err
	}

	// 将业务响应转换为 gRPC 响应
	return &grpc_generated.ReportDeviceIPResponse{
		DeviceCode: response.DeviceCode,
		Msg:        response.Msg,
	}, nil
}

// GetDevicePortForwards 实现获取设备端口转发的 gRPC 方法
func (s *OpenServer) GetDevicePortForwards(ctx context.Context, req *grpc_generated.GetDevicePortForwardsRequest) (*grpc_generated.GetDevicePortForwardsResponse, error) {
	// 调用你现有的业务逻辑（与 HTTP 处理器使用相同的代码）
	response, err := s.deviceService.DevicePortForwards(ctx, req.DeviceCode)
	if err != nil {
		return nil, err
	}

	// 将业务响应转换为 gRPC 响应
	var rules []*grpc_generated.PortForwardRule
	for _, rule := range response {
		rules = append(rules, &grpc_generated.PortForwardRule{
			Id:          rule.ID,
			LocalPort:   int32(rule.LocalPort),
			RemoteHost:  rule.RemoteHost,
			RemotePort:  int32(rule.RemotePort),
			Protocols:   rule.Protocols,
			AutoForward: rule.AutoForward,
		})
	}

	return &grpc_generated.GetDevicePortForwardsResponse{
		Rules: rules,
	}, nil
}
