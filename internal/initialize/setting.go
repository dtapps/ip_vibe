package initialize

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cache"
	"vibe_ip/internal/global"

	"github.com/spf13/cast"
)

// 初始化设置
func InitSetting(ctx context.Context, db *ent.Client) {
	// 定义一个内部函数简化读取逻辑
	getVal := func(key string, defaultVal string) string {
		c, err := db.Cache.Query().
			Where(cache.KeyEQ(key)).
			Only(ctx)
		if err != nil {
			// 如果没找到，就地创建默认值，保证下次直接命中
			_ = db.Cache.Create().
				SetKey(key).
				SetValue([]byte(defaultVal)).
				Exec(ctx)
			return defaultVal
		}
		return string(c.Value)
	}

	// 逐个填充 global.Setting
	global.GvaSetting = &global.Setting{
		// 设备离线失效时间（分，默认 12 分钟）
		DeviceOfflineInvalidTime: cast.ToInt64(getVal("conf:device_offline_invalid_time", "12")),
		// IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
		IPReportIntervalTime: getVal("conf:ip_report_interval_time", "8-12"),
		// 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
		PortForwardReloadTime: getVal("conf:port_reload_time", "30"),
		// 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
		PortForwardReportTime: getVal("conf:port_forward_report_time", "8-12"),
		// CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
		CDNRefreshTaskTime: getVal("conf:cdn_refresh_task_time", "10-12"),
		// DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
		DNSRefreshTaskTime: getVal("conf:dns_refresh_task_time", "10-12"),
		// SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）
		SSLRefreshTaskTime: getVal("conf:ssl_refresh_task_time", "8-10"),
	}
}
