package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Webhook(c *gin.Context) {
	// 请求参数
	var requestParams types.WebhookRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.WebhookResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetWebhookResponse, types.WebhookRequest]{
			Filter: requestParams,
			List:   []types.GetWebhookResponse{},
		},
	}

	// 创建服务
	service := service.NewWebhookService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaEnv.DEBUG,
	)

	// 查询全部通知
	allWebhook, err := global.GvaEntDb.Webhook.Query().
		WhereIf(requestParams.ID > 0, webhook.IDGT(requestParams.ID)).
		WhereIf(requestParams.Name != "", webhook.NameContains(requestParams.Name)).
		WhereIf(requestParams.Platform != "", webhook.PlatformEQ(webhook.Platform(requestParams.Platform))).
		WhereIf(!requestParams.IsActive.IsEmpty(), webhook.IsActiveEQ(requestParams.IsActive.Value())).
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allWebhook.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allWebhook.Page
	pageData.PageSize = allWebhook.Size
	pageData.Total = allWebhook.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "webhook.html"),
		types.BasePageData[types.WebhookResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "Webhook记录",
			Data:  pageData,
		},
	)
}
