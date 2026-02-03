package core

import (
	"context"
	"log/slog"
	"vibe_ip/internal/helper/ip_helper"
	"vibe_ip/internal/helper/sys_env"
)

// 获取IP数据 响应参数
type GetIPDataResult struct {
	IPV4Addr string // IP地址V4
	IPV6Addr string // IP地址V6
}

// GetIPData 获取IP数据
func GetIPData(ctx context.Context, logger *slog.Logger) (resp *GetIPDataResult, err error) {

	// 初始化参数
	resp = &GetIPDataResult{}

	// 获取设备IP
	iPHelper := ip_helper.NewIPHelper(logger, sys_env.GvaEnv.DEBUG)
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
