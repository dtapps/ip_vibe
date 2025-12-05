package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/ent"
	"vibe_ip/ent/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Webhook(c *gin.Context) {
	// 请求参数
	var requestParams types.WebhookRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 准备页面数据
	pageData := types.WebhookResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetWebhookResponse]{
			List: []types.GetWebhookResponse{},
		},
	}

	// 查询全部通知
	allWebhook, err := global.GvaEntDb.Webhook.Query().
		Order(ent.Desc(webhook.FieldCreatedAt)).
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allWebhook.Items {
			pageData.List = append(pageData.List, types.GetWebhookResponse{
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
	c.HTML(http.StatusOK, filepath.Join("pages", "webhook.html"), pageData)
}
