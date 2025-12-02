package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Webhook(c *gin.Context) {
	// 请求参数
	var requestParams types.WebhookRequest
	if err := gin_param.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

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
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allWebhook.Items {
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

	pageData.Page = allWebhook.Page
	pageData.PageSize = allWebhook.Size
	pageData.Total = allWebhook.Total

	// HTTP2 server 推送
	pusher(c)

	if global.GvaEnv.DEBUG {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, pageData)
	}

	// 渲染模板
	c.HTML(http.StatusOK, "pages/webhook.html", pageData)
}
