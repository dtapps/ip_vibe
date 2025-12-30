package handlers

import (
	"net/http"
	"path/filepath"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func WebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.WebhookSubRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.WebhookSubResponse{
		CommonPageResponse: types.CommonPageResponse[types.GetWebhookSubResponse, types.WebhookSubRequest]{
			Filter: requestParams,
			List:   []types.GetWebhookSubResponse{},
		},
	}

	// 创建服务
	service := service.NewWebhookSubService(
		global.GvaEntDb,
		global.GvaEnv.DEBUG,
		global.GvaOperationLogHelper,
	)

	// 查询全部通知订阅
	allWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		WithWebhook().
		OrderByCreatedAtDesc().
		Paginate(c, requestParams.Page, requestParams.PageSize)
	if err == nil {
		for _, item := range allWebhookSubscription.Items {
			list, err := service.BuildResponse(c, item)
			if err != nil {
				errorResponse(c, http.StatusNotFound, err.Error())
				return
			}
			pageData.List = append(pageData.List, *list)
		}
	}

	pageData.Page = allWebhookSubscription.Page
	pageData.PageSize = allWebhookSubscription.Size
	pageData.Total = allWebhookSubscription.Total

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "webhook-sub.html"),
		types.BasePageData[types.WebhookSubResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "Webhook订阅记录",
			Data:  pageData,
		},
	)
}
