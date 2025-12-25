package service

import (
	"context"
	"log/slog"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
)

// PortForwardServer 端口转发服务
type PortForwardServer struct {
	logger *slog.Logger // 日志记录器
}

// NewPortForwardServer 创建端口转发服务
func NewPortForwardServer(logger *slog.Logger) *PortForwardServer {
	return &PortForwardServer{
		logger: logger.With("component", "port-forward-server"),
	}
}

// GetDeviceData 获取设备数据
func (s *PortForwardServer) GetDeviceData(ctx context.Context, deviceCode, deviceName string) (resp *types.DeviceReportDeviceIpRequest, err error) {

	// 请求参数
	resp = &types.DeviceReportDeviceIpRequest{
		DevicePortForwardsRequest: types.DevicePortForwardsRequest{
			DeviceCode: deviceCode,
		},
		DeviceName: deviceName,
	}

	// 检查设备代码是否为空
	if resp.DeviceCode == "" {
		// 获取设备ID
		resp.DeviceCode, err = machineid.ID()
		if err != nil {
			s.logger.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}

	// 获取设备IP
	iPHelper := ip_helper.NewIPHelper()
	ipv4, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV4)
	if err == nil {
		resp.IPV4Addr = ipv4.IP
	}
	ipv6, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV6)
	if err == nil {
		resp.IPV6Addr = ipv6.IP
	}

	return resp, nil
}
