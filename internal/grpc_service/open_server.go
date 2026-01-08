package grpc_service

import (
	"context"
	"log/slog"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent_log/deviceipreport"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/types"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OpenServer 实现 gRPC 服务接口
type OpenServer struct {
	grpc_generated.UnimplementedOpenServiceServer                           // 嵌入未实现的服务
	logger                                        *slog.Logger              // 日志记录器
	deviceService                                 service.DeviceOpenService // 设备服务
}

// NewOpenServer 创建 gRPC 服务实例
func NewOpenServer(logger *slog.Logger) *OpenServer {
	// 使用与你的 HTTP 处理器相同的服务实例
	deviceService := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	return &OpenServer{
		logger:        logger,
		deviceService: deviceService,
	}
}

// RegisterService 注册服务到 gRPC
func (s *OpenServer) RegisterService(server *grpc.Server) {
	grpc_generated.RegisterOpenServiceServer(server, s)
}

// ReportDeviceIP 实现 设备上报 IP 方法
func (s *OpenServer) ReportDeviceIP(ctx context.Context, req *grpc_generated.ReportDeviceIPRequest) (*grpc_generated.ReportDeviceIPResponse, error) {
	// 将 gRPC 请求转换为你的 types 结构
	request := &types.DeviceReportDeviceIpRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		DeviceName: req.DeviceName,
		IPV4Addr:   req.Ipv4Addr,
		IPV6Addr:   req.Ipv6Addr,
		ReportType: deviceipreport.ReportTypeGrpc.String(),
	}

	// 调用服务方法
	response, err := s.deviceService.DeviceReportDeviceIp(ctx, request)
	if err != nil {
		return nil, err
	}

	// 转换为 gRPC 响应
	return &grpc_generated.ReportDeviceIPResponse{
		Msg: response.Msg,
	}, nil
}

// GetDevicePortForwards 实现 获取设备端口转发 方法
func (s *OpenServer) GetDevicePortForwards(ctx context.Context, req *grpc_generated.GetDevicePortForwardsRequest) (*grpc_generated.GetDevicePortForwardsResponse, error) {
	// 调用服务方法
	response, err := s.deviceService.DevicePortForwards(ctx, req.DeviceCode)
	if err != nil {
		return nil, err
	}

	// 转换为 gRPC 响应
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

	// 转换为 gRPC 响应
	return &grpc_generated.GetDevicePortForwardsResponse{
		Rules: rules,
	}, nil
}

// CreatePortForwardStats 实现 创建端口转发统计 方法
func (s *OpenServer) CreatePortForwardStats(ctx context.Context, req *grpc_generated.CreatePortForwardStatsRequest) (*emptypb.Empty, error) {
	// 将 gRPC 请求转换为你的 types 结构
	request := &types.CreatePortForwardStatsRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		PortForwardID: req.PortForwardId,
		StartTime:     req.StartTime,
		ActiveConn:    req.ActiveConn,
		TotalConn:     req.TotalConn,
		BytesIn:       req.BytesIn,
		BytesOut:      req.BytesOut,
		UDPPacketsIn:  req.UdpPacketsIn,
		UDPPacketsOut: req.UdpPacketsOut,
		Errors:        req.Errors,
	}

	// 调用服务方法
	err := s.deviceService.CreatePortForwardStats(ctx, req.DeviceCode, req.PortForwardId, request)
	if err != nil {
		return nil, err
	}

	// 转换为 gRPC 响应
	return &emptypb.Empty{}, nil
}
