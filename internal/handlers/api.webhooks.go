package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/webhook"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部通知
func GetWebhooks(c *gin.Context) {
	webhooks := make([]types.GetWebhookResponse, 0)
	// 查询全部通知
	allWebhook, err := global.GvaEntDb.Webhook.Query().
		Order(ent.Desc(webhook.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	for _, item := range allWebhook {
		webhooks = append(webhooks, types.GetWebhookResponse{
			ID:              item.ID,
			Name:            item.Name,
			Description:     item.Description,
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
	var requestParams types.GetWebhookRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知
	firstWebhook, err := global.GvaEntDb.Webhook.Query().
		Where(webhook.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	successResponse(c, types.GetWebhookResponse{
		ID:              firstWebhook.ID,
		Name:            firstWebhook.Name,
		Description:     firstWebhook.Description,
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
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建通知
	saveWebhook, err := global.GvaEntDb.Webhook.Create().
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
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
	var requestParams types.UpdateWebhookRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知
	firstWebhook, err := global.GvaEntDb.Webhook.Query().
		Where(webhook.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 更新通知
	saveWebhook, err := global.GvaEntDb.Webhook.UpdateOne(firstWebhook).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
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
	var requestParams types.UpdateWebhookActiveRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知
	firstWebhook, err := global.GvaEntDb.Webhook.Query().
		Where(webhook.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 更新通知状态
	saveWebhook, err := global.GvaEntDb.Webhook.UpdateOne(firstWebhook).
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
	var requestParams types.DeleteWebhookRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知
	firstWebhook, err := global.GvaEntDb.Webhook.Query().
		Where(webhook.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知失败: "+err.Error())
		return
	}
	// 删除通知
	err = global.GvaEntDb.Webhook.DeleteOne(firstWebhook).Exec(c)
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
