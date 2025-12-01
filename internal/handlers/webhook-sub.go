package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/device"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func WebhookSub(c *gin.Context) {
	// 准备页面数据
	pageData := types.WebhookSubResponse{
		List:     []types.GetWebhookSubscriptionsResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
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
			pageData.List = append(pageData.List, list)
		}
	}

	if global.GvaEnv.DEBUG {
		fmt.Printf("webhook-sub pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/webhook-sub.html", pageData)
}
