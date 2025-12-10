package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhook"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部通知
//
//	@Description	获取全部通知
//	@Summary		获取全部通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=[]types.GetWebhookResponse}	"成功响应"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/webhooks [get]
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
//
//	@Description	获取通知
//	@Summary		获取通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"通知序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数响应"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/webhooks/{id} [get]
func GetWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
//
//	@Description	创建通知
//	@Summary		创建通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateWebhookRequest					true	"Json 请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误响应"
//	@Router			/webhooks [post]
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
	successResponse(c, types.GetWebhookRequest{
		ID: saveWebhook.ID,
	})
}

// 更新通知
//
//	@Description	更新通知
//	@Summary		更新通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"通知序号"
//	@Param			request	body		types.UpdateWebhookRequest					true	"Json 请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误响应"
//	@Router			/webhooks/{id} [patch]
func UpdateWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
	successResponse(c, types.GetWebhookRequest{
		ID: saveWebhook.ID,
	})
}

// 更新通知启用状态
//
//	@Description	更新通知启用状态
//	@Summary		更新通知启用状态
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"通知序号"
//	@Param			request	body		types.UpdateWebhookActiveRequest			true	"Json 请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"参数响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Failure		500		{object}	APIResponse{}								"服务错误响应"
//	@Router			/webhooks/{id}/active [patch]
func UpdateWebhookActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
	successResponse(c, types.GetWebhookRequest{
		ID: saveWebhook.ID,
	})
}

// 删除通知
//
//	@Description	删除通知
//	@Summary		删除通知
//	@Tags			通知
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"通知序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数响应"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误响应"
//	@Router			/webhooks/{id} [delete]
func DeleteWebhook(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteWebhookRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
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
	successResponse(c, types.GetWebhookRequest{
		ID: firstWebhook.ID,
	})
}
