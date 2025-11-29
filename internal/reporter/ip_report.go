package reporter

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
)

// 全局互斥锁
var globalMutex sync.Mutex

func (s *Service) runTask() {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	var err error
	ctx := context.Background()

	// 请求参数
	deviceCode := global.GvaEnv.DEVICE_CODE
	params := types.DeviceReportDeviceIpRequest{
		DeviceName: global.GvaEnv.DEVICE_NAME,
	}
	// 检查设备代码是否为空
	if deviceCode == "" {
		// 获取设备ID
		deviceCode, err = machineid.ID()
		if err != nil {
			s.log.ErrorContext(ctx, "获取设备ID失败",
				slog.Any("err", err),
			)
			return
		}
	}

	// 获取设备IP
	iPHelper := ip_helper.NewIPHelper()
	ipv4, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV4)
	if err == nil {
		params.IPV4Addr = ipv4.IP
	}
	ipv6, err := iPHelper.GetIP(ctx, ip_helper.IPTypeV6)
	if err == nil {
		params.IPV6Addr = ipv6.IP
	}
	s.log.InfoContext(ctx, "IP上报开始",
		slog.String("device_code", deviceCode),
		slog.String("device_name", params.DeviceName),
		slog.String("ipv4_addr", params.IPV4Addr),
		slog.String("ipv6_addr", params.IPV6Addr),
	)
	err = s.httpApi(ctx, deviceCode, params)
	if err != nil {
		s.log.ErrorContext(ctx, "IP上报失败",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.Any("err", err),
		)
		return
	}

	nextRun, err := s.job.NextRun()
	if err == nil {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
			slog.String("next_run_time", nextRun.Format(time.DateTime)),
		)
	} else {
		s.log.InfoContext(ctx, "IP上报成功",
			slog.String("device_code", deviceCode),
			slog.String("device_name", params.DeviceName),
			slog.String("ipv4_addr", params.IPV4Addr),
			slog.String("ipv6_addr", params.IPV6Addr),
		)
	}

}
