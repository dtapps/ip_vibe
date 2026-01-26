package handlers

import (
	"log/slog"
	"path/filepath"
	"time"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/system_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	// 准备页面数据
	pageData := types.DashboardResponse{
		Stats:      []types.DashboardStat{},
		Entrances:  []types.DashboardEntrance{},
		RecentLogs: []types.GetLogResponse{},
	}

	// 设备总数
	deviceCount, _ := global.GvaEntDb.Device.Query().Count(c)
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "设备总数",
		Value: deviceCount,
	})

	// 在线设备总数 (心跳时间在 12 分钟内)
	invalidTime := db_cache.GvaSetting.DeviceOfflineInvalidTime // 单位：分钟
	onlineDeviceCount := 0
	if invalidTime <= 0 {
		onlineDeviceCount, _ = global.GvaEntDb.Device.Query().
			Count(c)
	} else {
		onlineDeviceCount, _ = global.GvaEntDb.Device.Query().
			Where(device.LastHeartbeatAtGTE(time.Now().Add(-time.Duration(invalidTime) * time.Minute))).
			Count(c)
	}
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "在线设备",
		Value: onlineDeviceCount,
	})

	// DNS 记录
	dnsRecordCount, _ := global.GvaEntDb.DNSRecord.Query().Count(c)
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "DNS 记录",
		Value: dnsRecordCount,
	})

	// CDN 记录
	cdnRecordCount, _ := global.GvaEntDb.CDNRecord.Query().Count(c)
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "CDN 记录",
		Value: cdnRecordCount,
	})

	// SSL 记录
	sslRecordCount, _ := global.GvaEntDb.SSLRecord.Query().Count(c)
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "SSL 记录",
		Value: sslRecordCount,
	})

	// 运行任务
	activeTaskCount := len(gocron_helper.NewJobHelper(global.GvaGoCron, global.GvaGocronSlog).GetJobsData(5))
	pageData.Stats = append(pageData.Stats, types.DashboardStat{
		Label: "运行任务",
		Value: activeTaskCount,
	})

	// 快速入口
	pageData.Entrances = []types.DashboardEntrance{
		{
			Title: "平台配置",
			Icon:  "i-lucide-server-cog",
			Href:  "/platform-config",
		},
		{
			Title: "CDN 管理",
			Icon:  "i-lucide-layers",
			Href:  "/cdn",
		},
		{
			Title: "DNS 管理",
			Icon:  "i-lucide-database-zap",
			Href:  "/dns",
		},
		{
			Title: "SSL 管理",
			Icon:  "i-lucide-shield-check",
			Href:  "/ssl",
		},
		{
			Title: "域名监控",
			Icon:  "i-lucide-activity",
			Href:  "/domain-monitor",
		},
		{
			Title: "设备管理",
			Icon:  "i-lucide-smartphone",
			Href:  "/device",
		},
		{
			Title: "IP 组",
			Icon:  "i-lucide-folder-plus",
			Href:  "/ip-group",
		},
		{
			Title: "端口转发",
			Icon:  "i-lucide-route",
			Href:  "/port-forward",
		},
		{
			Title: "通知管理",
			Icon:  "i-lucide-bell",
			Href:  "/webhook",
		},
		{
			Title: "任务管理",
			Icon:  "i-lucide-list-todo",
			Href:  "/task",
		},
	}

	// 最新操作日志
	allOperationLog, err := global.GvaEntLogDb.OperationLog.Query().
		Where(operationlog.SourceEQ(operationlog.SourceUser)).
		OrderByCreatedAtDesc().
		Limit(10).
		All(c)
	if err == nil {
		for _, item := range allOperationLog {
			pageData.RecentLogs = append(pageData.RecentLogs, types.GetLogResponse{
				ID:        item.ID,
				OpType:    item.OpType.String(),
				ObjID:     item.ObjID,
				ObjType:   item.ObjType.String(),
				OldValue:  item.OldValue,
				NewValue:  item.NewValue,
				Status:    item.Status.String(),
				Message:   item.Message,
				CreatedAt: item.CreatedAt,
			})
		}
	} else {
		slog.ErrorContext(c, "查询操作日志失败",
			slog.Any("err", err),
		)
	}

	// 获取自身设备ID
	pageData.DeviceCode, _ = machineid.ID()

	// 获取系统版本信息
	systemHelper := system_helper.NewSystemHelper()
	pageData.SystemVersion = global.GvaVersion
	pageData.SystemGitCommit = global.GvaGitCommit
	pageData.SystemBuildTime = global.GvaBuildTime
	pageData.SystemGoVersion = global.GvaGoVersion
	pageData.SystemOS = global.GvaOS
	pageData.SystemArch = global.GvaArch

	// 获取系统内存信息
	pageData.SystemMemoryInfo = systemHelper.GetMemoryInfo()

	// 获取系统设置
	// setting := db_cache.GetSetting(c, global.GvaEntDb)
	setting := db_cache.GvaSetting
	pageData.Settings = types.GetSettingResponse{
		DeviceOfflineInvalidTime: setting.DeviceOfflineInvalidTime,
		IPReportIntervalTime:     setting.IPReportIntervalTime,
		PortForwardReloadTime:    setting.PortForwardReloadTime,
		PortForwardReportTime:    setting.PortForwardReportTime,
		CDNRefreshTaskTime:       setting.CDNRefreshTaskTime,
		DNSRefreshTaskTime:       setting.DNSRefreshTaskTime,
		SSLApplyTaskTime:         setting.SSLApplyTaskTime,
		SSLRefreshTaskTime:       setting.SSLRefreshTaskTime,
		HTTPRequestLogRecordTime: setting.HTTPRequestLogRecordTime,
		HistoryDataRetainTime:    setting.HistoryDataRetainTime,
		StatisticDataRetainTime:  setting.StatisticDataRetainTime,
	}

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "dashboard.html"),
		types.BasePageData[types.DashboardResponse]{
			Title: "仪表盘",
			Data:  pageData,
		},
	)
}
