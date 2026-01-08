package dubbo_service

import (
	"context"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/dubbo_generated"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"dubbo.apache.org/dubbo-go/v3/server"
)

// OpenServer 实现 Dubbo 服务接口
type OpenServer struct {
	deviceService service.DeviceOpenService
}

// NewOpenServer 创建 Dubbo 服务实例
func NewOpenServer() *OpenServer {
	// 使用与你的 HTTP 处理器相同的服务实例
	deviceService := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	return &OpenServer{
		deviceService: deviceService,
	}
}

// RegisterService 注册服务到 Dubbo
func (s *OpenServer) RegisterService(server *server.Server) error {
	return dubbo_generated.RegisterOpenServiceHandler(server, s)
}

// ReportDeviceIP 实现 设备上报 IP 方法
func (s *OpenServer) ReportDeviceIP(ctx context.Context, req *dubbo_generated.ReportDeviceIPRequest) (*dubbo_generated.ReportDeviceIPResponse, error) {
	// 将 Dubbo 请求转换为你的 types 结构
	request := &types.DeviceReportDeviceIpRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		DeviceName: req.DeviceName,
		IPV4Addr:   req.Ipv4Addr,
		IPV6Addr:   req.Ipv6Addr,
	}

	// 调用服务方法
	response, err := s.deviceService.DeviceReportDeviceIp(ctx, request)
	if err != nil {
		return nil, err
	}

	// 转换为 Dubbo 响应
	return &dubbo_generated.ReportDeviceIPResponse{
		Msg: response.Msg,
	}, nil
}

// GetDevicePortForwards 实现 获取设备端口转发 方法
func (s *OpenServer) GetDevicePortForwards(ctx context.Context, req *dubbo_generated.GetDevicePortForwardsRequest) (*dubbo_generated.GetDevicePortForwardsResponse, error) {
	// 调用服务方法
	response, err := s.deviceService.DevicePortForwards(ctx, req.DeviceCode)
	if err != nil {
		return nil, err
	}

	// 转换为 Dubbo 响应
	var rules []*dubbo_generated.PortForwardRule
	for _, rule := range response {
		rules = append(rules, &dubbo_generated.PortForwardRule{
			Id:          rule.ID,
			LocalPort:   int32(rule.LocalPort),
			RemoteHost:  rule.RemoteHost,
			RemotePort:  int32(rule.RemotePort),
			Protocols:   rule.Protocols,
			AutoForward: rule.AutoForward,
		})
	}

	// 转换为 Dubbo 响应
	return &dubbo_generated.GetDevicePortForwardsResponse{
		Rules: rules,
	}, nil
}

// CreatePortForwardStats 实现 创建端口转发统计 方法
func (s *OpenServer) CreatePortForwardStats(ctx context.Context, req *dubbo_generated.CreatePortForwardStatsRequest) (*dubbo_generated.CreatePortForwardStatsResponse, error) {
	// 将 Dubbo 请求转换为你的 types 结构
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

	// 转换为 Dubbo 响应
	return &dubbo_generated.CreatePortForwardStatsResponse{}, nil
}
