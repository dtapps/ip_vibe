package db_cache

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cache"

	"github.com/spf13/cast"
)

// Setting 设置
type Setting struct {
	// 设备离线失效时间（分，默认 12 分钟）
	DeviceOfflineInvalidTime int64 `json:"device_offline_invalid_time"`
	// IP上报间隔时间（分，默认 8-12 分钟，范围时用-分隔）
	IPReportIntervalTime string `json:"ip_report_interval_time"`
	// 端口转发重载时间（分，默认 30 分钟，范围时用-分隔）
	PortForwardReloadTime string `json:"port_forward_reload_time"`
	// 端口转发统计上报时间（分，默认 8-12 分钟，范围时用-分隔）
	PortForwardReportTime string `json:"port_forward_report_time"`
	// CDN刷新任务时间（分，默认 10-12 分钟，范围时用-分隔）
	CDNRefreshTaskTime string `json:"cdn_refresh_task_time"`
	// DNS刷新任务时间（分，默认 10-12 分钟，范围时用-分隔）
	DNSRefreshTaskTime string `json:"dns_refresh_task_time"`
	// TUN刷新任务时间（分，默认 10-12 分钟，范围时用-分隔）
	TUNRefreshTaskTime string `json:"tun_refresh_task_time"`
	// SSL申请任务时间（时，默认 8-10 小时，范围时用-分隔）
	SSLApplyTaskTime string `json:"ssl_apply_task_time"`
	// SSL刷新任务时间（时，默认 12-24 小时，范围时用-分隔）
	SSLRefreshTaskTime string `json:"ssl_refresh_task_time"`
	// CloudFlare Tunnel同步任务时间（时，默认 12-24 小时，范围时用-分隔）
	CloudflareTunnelSyncTaskTime string `json:"cloudflare_tunnel_sync_task_time"`

	// HTTP请求日志记录时间（天，默认 90 天）
	HTTPRequestLogRecordTime int64 `json:"http_request_log_record_time"`
	// 历史数据保留时间（天，默认 365 天）
	HistoryDataRetainTime int64 `json:"history_data_retain_time"`
	// 统计数据保留时间（天，默认 365 天）
	StatisticDataRetainTime int64 `json:"statistic_data_retain_time"`
}

// GvaSetting 全局设置
var GvaSetting *Setting

// 统一管理所有 Cache Key
const (
	KeyDeviceOfflineTime        = "conf:device_offline_invalid_time"
	KeyIPReportInterval         = "conf:ip_report_interval_time"
	KeyPortForwardReload        = "conf:port_forward_reload_time"
	KeyPortForwardReport        = "conf:port_forward_report_time"
	KeyCDNRefreshTime           = "conf:cdn_refresh_task_time"
	KeyDNSRefreshTime           = "conf:dns_refresh_task_time"
	KeyTUNRefreshTime           = "conf:tun_refresh_task_time"
	KeySSLApplyTime             = "conf:ssl_apply_task_time"
	KeySSLRefreshTime           = "conf:ssl_refresh_task_time"
	KeyCloudflareTunnelSyncTime = "conf:cloudflare_tunnel_sync_task_time"

	KeyHTTPRequestLogRecordTime = "conf:http_request_log_record_time"
	KeyHistoryDataRetainTime    = "conf:history_data_retain_time"
	KeyStatisticDataRetainTime  = "conf:statistic_data_retain_time"
)

// 初始化设置
func InitSetting(ctx context.Context, db *ent.Client) {
	// 内部辅助函数：查不到就建默认的
	getVal := func(key string, defaultVal string) string {
		c, err := db.Cache.Query().
			Where(cache.KeyEQ(key)).
			Only(ctx)
		if err != nil {
			// 如果数据库里没有，就地初始化一条记录
			_ = db.Cache.Create().
				SetKey(key).
				SetValue([]byte(defaultVal)).
				Exec(ctx)
			return defaultVal
		}
		return string(c.Value)
	}

	// 设置变量
	GvaSetting = new(Setting)
	GvaSetting.DeviceOfflineInvalidTime = cast.ToInt64(getVal(KeyDeviceOfflineTime, "12"))
	GvaSetting.IPReportIntervalTime = getVal(KeyIPReportInterval, "8-12")
	GvaSetting.PortForwardReloadTime = getVal(KeyPortForwardReload, "30")
	GvaSetting.PortForwardReportTime = getVal(KeyPortForwardReport, "8-12")
	GvaSetting.CDNRefreshTaskTime = getVal(KeyCDNRefreshTime, "10-12")
	GvaSetting.DNSRefreshTaskTime = getVal(KeyDNSRefreshTime, "10-12")
	GvaSetting.TUNRefreshTaskTime = getVal(KeyTUNRefreshTime, "10-12")
	GvaSetting.SSLApplyTaskTime = getVal(KeySSLApplyTime, "8-10")
	GvaSetting.SSLRefreshTaskTime = getVal(KeySSLRefreshTime, "12-24")
	GvaSetting.CloudflareTunnelSyncTaskTime = getVal(KeyCloudflareTunnelSyncTime, "12-24")

	GvaSetting.HTTPRequestLogRecordTime = cast.ToInt64(getVal(KeyHTTPRequestLogRecordTime, "90"))
	GvaSetting.HistoryDataRetainTime = cast.ToInt64(getVal(KeyHistoryDataRetainTime, "365"))
	GvaSetting.StatisticDataRetainTime = cast.ToInt64(getVal(KeyStatisticDataRetainTime, "365"))

}

// 获取设置
func GetSetting(ctx context.Context, db *ent.Client) *Setting {
	// 内部辅助函数：查不到就建默认的
	getVal := func(key string, defaultVal string) string {
		c, err := db.Cache.Query().
			Where(cache.KeyEQ(key)).
			Only(ctx)
		if err != nil {
			// 如果数据库里没有，就地初始化一条记录
			_ = db.Cache.Create().
				SetKey(key).
				SetValue([]byte(defaultVal)).
				Exec(ctx)
			return defaultVal
		}
		return string(c.Value)
	}

	// 返回变量
	setting := new(Setting)
	setting.DeviceOfflineInvalidTime = cast.ToInt64(getVal(KeyDeviceOfflineTime, "12"))
	setting.IPReportIntervalTime = getVal(KeyIPReportInterval, "8-12")
	setting.PortForwardReloadTime = getVal(KeyPortForwardReload, "30")
	setting.PortForwardReportTime = getVal(KeyPortForwardReport, "8-12")
	setting.CDNRefreshTaskTime = getVal(KeyCDNRefreshTime, "10-12")
	setting.DNSRefreshTaskTime = getVal(KeyDNSRefreshTime, "10-12")
	setting.TUNRefreshTaskTime = getVal(KeyTUNRefreshTime, "10-12")
	setting.SSLApplyTaskTime = getVal(KeySSLApplyTime, "8-10")
	setting.SSLRefreshTaskTime = getVal(KeySSLRefreshTime, "12-24")
	setting.CloudflareTunnelSyncTaskTime = getVal(KeyCloudflareTunnelSyncTime, "12-24")

	setting.HTTPRequestLogRecordTime = cast.ToInt64(getVal(KeyHTTPRequestLogRecordTime, "90"))
	setting.HistoryDataRetainTime = cast.ToInt64(getVal(KeyHistoryDataRetainTime, "365"))
	setting.StatisticDataRetainTime = cast.ToInt64(getVal(KeyStatisticDataRetainTime, "365"))

	return setting
}

// 批量更新设置
func UpdateSetting(ctx context.Context, db *ent.Client, newSetting *Setting) error {
	// 定义 字段 -> Key 的映射关系，循环处理
	updates := map[string]any{
		KeyDeviceOfflineTime:        newSetting.DeviceOfflineInvalidTime,
		KeyIPReportInterval:         newSetting.IPReportIntervalTime,
		KeyPortForwardReload:        newSetting.PortForwardReloadTime,
		KeyPortForwardReport:        newSetting.PortForwardReportTime,
		KeyCDNRefreshTime:           newSetting.CDNRefreshTaskTime,
		KeyDNSRefreshTime:           newSetting.DNSRefreshTaskTime,
		KeyTUNRefreshTime:           newSetting.TUNRefreshTaskTime,
		KeySSLApplyTime:             newSetting.SSLApplyTaskTime,
		KeySSLRefreshTime:           newSetting.SSLRefreshTaskTime,
		KeyCloudflareTunnelSyncTime: newSetting.CloudflareTunnelSyncTaskTime,

		KeyHTTPRequestLogRecordTime: newSetting.HTTPRequestLogRecordTime,
		KeyHistoryDataRetainTime:    newSetting.HistoryDataRetainTime,
		KeyStatisticDataRetainTime:  newSetting.StatisticDataRetainTime,
	}

	// 开启 Upsert 模式批量写入 Cache 表
	for k, v := range updates {
		err := SaveSingleSetting(ctx, db, k, v)
		if err != nil {
			return fmt.Errorf("更新配置项 %s 失败: %w", k, err)
		}
	}

	// 数据库写完后，更新内存中的全局变量，保证即刻生效
	*GvaSetting = *newSetting
	return nil
}

// 保存单个配置
func SaveSingleSetting(ctx context.Context, db *ent.Client, key string, value any) error {
	return db.Cache.Create().
		SetKey(key).
		SetValue([]byte(cast.ToString(value))).
		OnConflictColumns(cache.FieldKey). // 依赖你的 Cache Schema 中 Key 字段的 Unique 索引
		UpdateValue().
		Exec(ctx)
}
