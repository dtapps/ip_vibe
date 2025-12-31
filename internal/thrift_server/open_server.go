package thrift_server

import (
	"context"
	"vibe_ip/internal/backend/service"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/thrift_generated/open"
	"vibe_ip/internal/types"

	"github.com/spf13/cast"
)

// 实现 Thrift 服务接口
type OpenService struct {
	deviceService service.DeviceOpenService
}

// NewOpenService 创建 Thrift 服务实例
func NewOpenService() *OpenService {
	// 使用与你的 HTTP 处理器相同的服务实例
	deviceService := service.NewDeviceOpenService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaOperationLogHelper,
		webhookService.NewSend(global.GvaEntDb, global.GvaEntLogDb),
		global.GvaEnv.DEBUG,
	)

	return &OpenService{
		deviceService: deviceService,
	}
}

// ReportDeviceIP 实现 设备上报 IP 方法
func (s *OpenService) ReportDeviceIP(ctx context.Context, req open.ReportDeviceIPRequest) (r open.ReportDeviceIPResponse, err error) {
	// 将 Thrift 请求转换为你的 types 结构
	request := &types.DeviceReportDeviceIpRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		DeviceName: cast.ToString(req.DeviceName),
		IPV4Addr:   cast.ToString(req.Ipv4Addr),
		IPV6Addr:   cast.ToString(req.Ipv6Addr),
	}

	// 调用服务方法
	response, err := s.deviceService.DeviceReportDeviceIp(ctx, request)
	if err != nil {
		return open.ReportDeviceIPResponse{}, err
	}

	// 转换为 Thrift 响应
	return open.ReportDeviceIPResponse{
		Msg: &response.Msg,
	}, nil
}

// GetDevicePortForwards 实现 获取设备端口转发 方法
func (s *OpenService) GetDevicePortForwards(ctx context.Context, req open.GetDevicePortForwardsRequest) (r open.GetDevicePortForwardsResponse, err error) {
	// 调用服务方法
	response, err := s.deviceService.DevicePortForwards(ctx, req.DeviceCode)
	if err != nil {
		return open.GetDevicePortForwardsResponse{}, err
	}

	// 转换为 Thrift 响应
	var rules []*open.PortForwardRule
	for _, rule := range response {
		rules = append(rules, &open.PortForwardRule{
			ID:          rule.ID,
			LocalPort:   int32(rule.LocalPort),
			RemoteHost:  rule.RemoteHost,
			RemotePort:  int32(rule.RemotePort),
			Protocols:   rule.Protocols,
			AutoForward: rule.AutoForward,
		})
	}

	// 转换为 Thrift 响应
	return open.GetDevicePortForwardsResponse{
		Rules: rules,
	}, nil
}

// CreatePortForwardStats 实现 创建端口转发统计 方法
func (s *OpenService) CreatePortForwardStats(ctx context.Context, req open.CreatePortForwardStatsRequest) (r open.CreatePortForwardStatsResponse, err error) {
	// 将 Thrift 请求转换为你的 types 结构
	request := &types.CreatePortForwardStatsRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: req.DeviceCode,
		},
		PortForwardID: req.PortForwardID,
		StartTime:     req.StartTime,
		ActiveConn:    cast.ToInt64(req.ActiveConn),
		TotalConn:     cast.ToInt64(req.TotalConn),
		BytesIn:       cast.ToInt64(req.BytesIn),
		BytesOut:      cast.ToInt64(req.BytesOut),
		UDPPacketsIn:  cast.ToInt64(req.UDPPacketsIn),
		UDPPacketsOut: cast.ToInt64(req.UDPPacketsOut),
		Errors:        cast.ToInt64(req.Errors),
	}

	// 调用服务方法
	err = s.deviceService.CreatePortForwardStats(ctx, req.DeviceCode, req.PortForwardID, request)
	if err != nil {
		return open.CreatePortForwardStatsResponse{}, err
	}

	// 转换为 Thrift 响应
	return open.CreatePortForwardStatsResponse{}, nil
}
