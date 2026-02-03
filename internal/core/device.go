package core

import (
	"context"
	"log/slog"
	"vibe_ip/internal/helper/sys_env"

	"github.com/denisbrodbeck/machineid"
)

// 获取设备数据 响应参数
type GetDeviceDataResult struct {
	DeviceCode string // 设备编码
	DeviceName string // 设备名称
}

// GetDeviceData 获取设备数据
func GetDeviceData(ctx context.Context, logger *slog.Logger) (resp *GetDeviceDataResult, err error) {

	// 初始化参数
	resp = &GetDeviceDataResult{
		DeviceCode: sys_env.GvaEnv.DEVICE_CODE,
		DeviceName: sys_env.GvaEnv.DEVICE_NAME,
	}

	// 检查设备代码是否为空
	if resp.DeviceCode == "" {
		// 获取设备ID
		resp.DeviceCode, err = machineid.ID()
		if err != nil {
			logger.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}

	return resp, nil
}
