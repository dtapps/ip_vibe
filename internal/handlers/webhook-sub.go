package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/device"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func WebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.WebhookSubRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.WebhookSubResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetWebhookSubResponse]{
			List: []types.GetWebhookSubResponse{},
		},
	}

	// 查询全部通知订阅
	allWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		WithWebhook().
		Order(ent.Desc(webhooksubscription.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allWebhookSubscription.Items {
			list := types.GetWebhookSubResponse{
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

	pageData.Page = allWebhookSubscription.Page
	pageData.PageSize = allWebhookSubscription.Size
	pageData.Total = allWebhookSubscription.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, filepath.Join("pages", "webhook-sub.html"), pageData)
}
