package service

import (
	"context"
	"log/slog"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
)

// IPReportService IP报告服务
type IPReportService struct {
	logger *slog.Logger // 日志记录器
}

// NewIPReportService 创建IP报告服务
func NewIPReportService(logger *slog.Logger) *IPReportService {
	return &IPReportService{
		logger: logger.With("component", "ip-report-service"),
	}
}

// GetDeviceData 获取设备数据
func (s *IPReportService) GetDeviceData(ctx context.Context, deviceCode, deviceName string) (resp *types.DeviceReportDeviceIpRequest, err error) {

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
