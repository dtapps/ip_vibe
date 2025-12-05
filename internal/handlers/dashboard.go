package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/system_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
)

func Dashboard(c *gin.Context) {
	// 准备页面数据
	pageData := types.DashboardResponse{
		RecentLogs: []types.GetLogResponse{},
	}
	pageData.Stats.LogLimit = 15

	// 设备总数
	pageData.Stats.DeviceCount, _ = global.GvaEntDb.Device.Query().Count(c)

	// 在线设备总数 (心跳时间在 12 分钟内)
	pageData.Stats.OnlineDeviceCount, _ = global.GvaEntDb.Device.Query().
		Where(device.LastHeartbeatAtGTE(time.Now().Add(-12 * time.Minute))).
		Count(c)

	// DNS 记录
	pageData.Stats.DnsRecordCount, _ = global.GvaEntDb.DNSRecord.Query().Count(c)

	// CDN 记录
	pageData.Stats.CdnRecordCount, _ = global.GvaEntDb.CDNRecord.Query().Count(c)

	// 通知渠道
	pageData.Stats.WebhookCount, _ = global.GvaEntDb.Webhook.Query().Count(c)

	// 运行任务
	pageData.Stats.ActiveTaskCount = len(gocron_helper.NewJobHelper(global.GvaGoCron).GetJobsData(5))

	// 最新操作日志
	allOperationLog, err := global.GvaEntLogDb.OperationLog.Query().
		Order(ent.Desc(operationlog.FieldCreatedAt)).
		Limit(pageData.Stats.LogLimit).
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

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "dashboard.html"), pageData)
}
