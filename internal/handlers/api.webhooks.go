package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/webhook"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部通知
func GetWebhooks(c *gin.Context) {
	webhooks := make([]types.GetWebhooksResponse, 0)
	// 查询全部通知
	allWebhook, err := global.GvaDb.Webhook.Query().
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	for _, item := range allWebhook {
		webhooks = append(webhooks, types.GetWebhooksResponse{
			ID:              item.ID,
			Name:            item.Name,
			Platform:        item.Platform.String(),
			URL:             item.URL,
			Secret:          item.Secret,
			Events:          item.Events,
			IsActive:        item.IsActive,
			TemplateTitle:   item.TemplateTitle,
			TemplateContent: item.TemplateContent,
		})
	}
	successResponse(c, webhooks)
}

// 获取通知
func GetWebhook(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询通知
	firstWebhook, err := global.GvaDb.Webhook.Query().
		Where(webhook.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	successResponse(c, types.GetWebhooksResponse{
		ID:              firstWebhook.ID,
		Name:            firstWebhook.Name,
		Platform:        firstWebhook.Platform.String(),
		URL:             firstWebhook.URL,
		Secret:          firstWebhook.Secret,
		Events:          firstWebhook.Events,
		IsActive:        firstWebhook.IsActive,
		TemplateTitle:   firstWebhook.TemplateTitle,
		TemplateContent: firstWebhook.TemplateContent,
	})
}

// 创建通知
func CreateWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateWebhookRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("CreateWebhook params: %+v\n", requestParams)
	}
	// 创建通知
	saveWebhook, err := global.GvaDb.Webhook.Create().
		SetName(requestParams.Name).
		SetPlatform(webhook.Platform(requestParams.Platform)).
		SetURL(requestParams.URL).
		SetSecret(requestParams.Secret).
		SetEvents(requestParams.Events).
		SetTemplateTitle(requestParams.TemplateTitle).
		SetTemplateContent(requestParams.TemplateContent).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建通知: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建通知失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "创建通知")
	successResponse(c, gin.H{
		"id": saveWebhook.ID,
	})
}

// 更新通知
func UpdateWebhook(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	var requestParams types.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("UpdateWebhook params: %+v\n", requestParams)
	}
	// 查询通知
	firstWebhook, err := global.GvaDb.Webhook.Query().
		Where(webhook.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 更新通知
	saveWebhook, err := global.GvaDb.Webhook.UpdateOne(firstWebhook).
		SetName(requestParams.Name).
		SetSecret(requestParams.Secret).
		SetEvents(requestParams.Events).
		SetTemplateTitle(requestParams.TemplateTitle).
		SetTemplateContent(requestParams.TemplateContent).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新通知: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新通知失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "更新通知")
	successResponse(c, gin.H{
		"id": saveWebhook.ID,
	})
}

// 更新通知状态
func UpdateWebhookActive(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	var requestParams types.UpdateWebhookActiveRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("UpdateWebhookActive params: %+v\n", requestParams)
	}
	// 查询通知
	firstWebhook, err := global.GvaDb.Webhook.Query().
		Where(webhook.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 更新通知状态
	saveWebhook, err := global.GvaDb.Webhook.UpdateOne(firstWebhook).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新通知状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新通知状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "更新通知状态")
	successResponse(c, gin.H{
		"id": saveWebhook.ID,
	})
}

// 删除通知
func DeleteWebhook(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询通知
	firstWebhook, err := global.GvaDb.Webhook.Query().
		Where(webhook.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 删除通知
	err = global.GvaDb.Webhook.DeleteOne(firstWebhook).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除通知: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除通知失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstWebhook.ID, operationlog.ObjTypeWebhook, operationlog.StatusSuccess, "删除通知")
	successResponse(c, gin.H{
		"id": firstWebhook.ID,
	})
}
