package service

import (
	"context"
	"log/slog"

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

// 获取设备数据 响应参数
type GetDeviceDataResult struct {
	DeviceCode string // 设备编码
	DeviceName string // 设备名称
}

// GetDeviceData 获取设备数据
func (s *PortForwardServer) GetDeviceData(ctx context.Context, deviceCode, deviceName string) (resp *GetDeviceDataResult, err error) {

	// 请求参数
	resp = &GetDeviceDataResult{
		DeviceCode: deviceCode,
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

	return resp, nil
}
