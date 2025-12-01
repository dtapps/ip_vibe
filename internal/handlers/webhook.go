package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Webhook(c *gin.Context) {
	// 准备页面数据
	pageData := types.WebhookResponse{
		List:     []types.GetWebhooksResponse{},
		Page:     0,
		PageSize: 0,
		Total:    0,
	}

	// 查询全部通知
	allWebhook, err := global.GvaDb.Webhook.Query().
		Order(ent.Desc(webhook.FieldCreatedAt)).
		All(c)
	if err == nil {
		for _, item := range allWebhook {
			pageData.List = append(pageData.List, types.GetWebhooksResponse{
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
	if global.GvaEnv.DEBUG {
		fmt.Printf("webhook pageData: %+v \n", pageData)
	}

	// HTTP2 server 推送
	pusher(c)

	// 渲染模板
	c.HTML(http.StatusOK, "pages/webhook.html", pageData)
}
