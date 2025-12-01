package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/device"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/ipgroup"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/platformconfig"
	"vibe_ip/ent/webhook"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/system_helper"
	"vibe_ip/internal/types"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
)

func Index(c *gin.Context) {
	// 准备页面数据
	pageData := types.IndexResponse{
		Configs:              []types.GetConfigsResponse{},
		CdnRecords:           []types.GetCdnRecordsResponse{},
		DnsRecords:           []types.GetDnsRecordsResponse{},
		Devices:              []types.GetDevicesResponse{},
		IpGroups:             []types.GetIPGroupsResponse{},
		PortForwards:         []types.GetPortForwardsResponse{},
		Webhooks:             []types.GetWebhooksResponse{},
		WebhookSubscriptions: []types.GetWebhookSubscriptionsResponse{},
		Tasks:                []gocron_helper.JobData{},

		OperationLogs:     []types.GetOperationLogsResponse{},
		OperationLogLimit: 0,
	}
	// 查询全部配置
	allConfig, err := global.GvaDb.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allConfig {
			pageData.Configs = append(pageData.Configs, types.GetConfigsResponse{
				ID:          item.ID,
				Capability:  item.Capability.String(),
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
			})
		}
	}
	// 查询全部CDN记录
	allCdnRecord, err := global.GvaDb.CDNRecord.Query().
		WithConfig().
		WithIPGroup().
		Order(ent.Desc(cdnrecord.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allCdnRecord {
			list := types.GetCdnRecordsResponse{
				ID:             item.ID,
				ConfigID:       item.ConfigID,
				Platform:       item.Edges.Config.Platform.String(),
				RecordID:       item.RecordID,
				Domain:         item.Domain,
				Type:           item.Type.String(),
				OriginProtocol: item.OriginProtocol.String(),
				HTTPPort:       item.HTTPPort,
				HTTPSPort:      item.HTTPSPort,
				CdnDomain:      item.CdnDomain,
				IsActive:       item.IsActive,
				LastRefreshAt:  item.LastRefreshAt,
			}
			if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
				list.IPGroupID = ipGroup.ID
				list.IPGroupName = ipGroup.Name
			}
			pageData.CdnRecords = append(pageData.CdnRecords, list)
		}
	}
	// 查询全部DNS记录
	allDnsRecord, err := global.GvaDb.DNSRecord.Query().
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		Order(ent.Desc(dnsrecord.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allDnsRecord {
			list := types.GetDnsRecordsResponse{
				ID:            item.ID,
				ConfigID:      item.ConfigID,
				Platform:      item.Edges.Config.Platform.String(),
				RecordID:      item.RecordID,
				Domain:        item.Domain,
				Type:          item.Type.String(),
				IsActive:      item.IsActive,
				LastRefreshAt: item.LastRefreshAt,
			}
			if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
				list.IPGroupID = ipGroup.ID
				list.IPGroupName = ipGroup.Name
			}
			if cdnRecord, err := item.Edges.CdnRecordOrErr(); err == nil {
				list.CdnRecordID = cdnRecord.ID
				list.CdnRecordCdnDomain = cdnRecord.CdnDomain
			}
			pageData.DnsRecords = append(pageData.DnsRecords, list)
		}
	}
	// 查询全部设备
	allDevice, err := global.GvaDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allDevice {
			pageData.Devices = append(pageData.Devices, types.GetDevicesResponse{
				ID:              item.ID,
				Code:            item.Code,
				Name:            item.Name,
				Description:     item.Description,
				IPV4:            item.IPV4,
				IPV6:            item.IPV6,
				LastReportedAt:  item.LastReportedAt,
				LastHeartbeatAt: item.LastHeartbeatAt,
			})
		}
	}
	// 查询全部IP组
	allIpGroup, err := global.GvaDb.IpGroup.Query().
		WithDevices().
		Order(ent.Desc(ipgroup.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allIpGroup {
			pageData.IpGroups = append(pageData.IpGroups, types.GetIPGroupsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				DeviceCount: len(item.Edges.Devices),
			})
		}
	}
	// 查询全部端口转发
	allPortForward, err := global.GvaDb.DevicePortForward.Query().
		WithDevice().
		Order(ent.Desc(deviceportforward.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allPortForward {
			pageData.PortForwards = append(pageData.PortForwards, types.GetPortForwardsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				LocalPort:   item.LocalPort,
				RemoteHost:  item.RemoteHost,
				RemotePort:  item.RemotePort,
				Protocols:   item.Protocols,
				AutoForward: item.AutoForward,
				IsActive:    item.IsActive,
				DeviceID:    item.Edges.Device.ID,
				DeviceName:  item.Edges.Device.Name,
			})
		}
	}
	// 查询全部通知
	allWebhook, err := global.GvaDb.Webhook.Query().
		Order(ent.Desc(webhook.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allWebhook {
			pageData.Webhooks = append(pageData.Webhooks, types.GetWebhooksResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				Platform:    item.Platform.String(),
				URL:         item.URL,
				Secret:      item.Secret,
				Events:      item.Events,
				IsActive:    item.IsActive,
			})
		}
	}
	// 查询全部通知订阅
	allWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		WithWebhook().
		Order(ent.Desc(webhooksubscription.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allWebhookSubscription {
			list := types.GetWebhookSubscriptionsResponse{
				ID:          item.ID,
				Name:        item.Name,
				Description: item.Description,
				WebhookID:   item.WebhookID,
				WebhookName: item.Edges.Webhook.Name,
				ObjType:     item.ObjType.String(),
				ObjID:       item.ObjID,
				IsActive:    item.IsActive,
			}
			switch item.ObjType {
			case webhooksubscription.ObjTypeCdnRecord:
				firstCdnRecord, err := global.GvaDb.CDNRecord.Query().
					Where(cdnrecord.IDEQ(item.ObjID)).
					First(c)
				if err == nil {
					list.ObjName = firstCdnRecord.Domain
				}
			case webhooksubscription.ObjTypeDNSRecord:
				firstDNSRecord, err := global.GvaDb.DNSRecord.Query().
					Where(dnsrecord.IDEQ(item.ObjID)).
					First(c)
				if err == nil {
					list.ObjName = firstDNSRecord.Domain
				}
			case webhooksubscription.ObjTypeDevice:
				firstDevice, err := global.GvaDb.Device.Query().
					Where(device.IDEQ(item.ObjID)).
					First(c)
				if err == nil {
					list.ObjName = firstDevice.Name
				}
			}
			pageData.WebhookSubscriptions = append(pageData.WebhookSubscriptions, list)
		}
	}
	// 任务助手
	jobHelper := gocron_helper.NewJobHelper(global.GvaGoCron)
	// 获取全部任务
	pageData.Tasks = jobHelper.GetJobsData(5)
	// 查询全部操作日志
	allOperationLog, err := global.GvaDb.OperationLog.Query().
		Order(ent.Desc(operationlog.FieldCreatedAt)).
		Limit(15).
		All(c)
	if err == nil {
		for _, item := range allOperationLog {
			pageData.OperationLogs = append(pageData.OperationLogs, types.GetOperationLogsResponse{
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
		pageData.OperationLogLimit = 15
	}

	// 查询任务状态
	if global.GvaGoCronRunning == 1 {
		pageData.TaskStatus = "running"
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

	// if global.GvaEnv.DEBUG {
	// 	fmt.Printf("Index pageData: %+v \n", pageData)
	// }

	// HTTP2 server 推送
	// https://gin-gonic.com/zh-cn/docs/examples/http2-server-push/
	if pusher := c.Writer.Pusher(); pusher != nil {
		// 推送 CSS 和 JS 文件
		resources := []string{
			"/static/index.css",
			"/static/index.dark.css",
			"/static/uuid.js",
			"/static/index.js",
		}
		for _, res := range resources {
			if err := pusher.Push(res, nil); err != nil {
				fmt.Printf("推送失败 %s: %v \n", res, err)
			}
		}
	} else {
		if global.GvaEnv.DEBUG {
			fmt.Printf("推送器为空，跳过推送资源 \n")
		}
	}

	// 渲染模板
	c.HTML(http.StatusOK, "index.html", pageData)
}
