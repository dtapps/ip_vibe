package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
	"vibe_ip/internal/ent/device"
	"vibe_ip/internal/ent/dnsrecord"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部通知订阅
//
//	@Description	获取全部通知订阅
//	@Summary		获取全部通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.GetWebhookSubsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetWebhookSubResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/webhook-subs [get]
func GetWebhookSubs(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookSubsRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	webhooks := make([]types.GetWebhookSubResponse, 0)
	// 查询全部通知订阅
	allWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		WhereIf(requestParams.IsActive != nil, webhooksubscription.IsActiveEQ(cast.ToBool(requestParams.IsActive))).
		Order(ent.Desc(webhooksubscription.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	for _, item := range allWebhookSubscription {
		list := types.GetWebhookSubResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			WebhookID:   item.WebhookID,
			ObjType:     item.ObjType.String(),
			ObjID:       item.ObjID,
			IsActive:    item.IsActive,
		}
		switch item.ObjType {
		case webhooksubscription.ObjTypeCdnRecord:
			firstCDNRecord, err := global.GvaEntDb.CDNRecord.Query().
				Where(cdnrecord.IDEQ(item.ObjID)).
				First(c)
			if err == nil {
				list.ObjName = firstCDNRecord.Domain
			}
		case webhooksubscription.ObjTypeDNSRecord:
			firstDNSRecord, err := global.GvaEntDb.DNSRecord.Query().
				Where(dnsrecord.IDEQ(item.ObjID)).
				First(c)
			if err == nil {
				list.ObjName = firstDNSRecord.Domain
			}
		case webhooksubscription.ObjTypeDevice:
			firstDevice, err := global.GvaEntDb.Device.Query().
				Where(device.IDEQ(item.ObjID)).
				First(c)
			if err == nil {
				list.ObjName = firstDevice.Name
			}
		}
		webhooks = append(webhooks)
	}
	successResponse(c, webhooks)
}

// 获取通知订阅
//
//	@Description	获取通知订阅
//	@Summary		获取通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"通知订阅序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookSubResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"请求参数"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/webhook-subs/{id} [get]
func GetWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	list := types.GetWebhookSubResponse{
		ID:          firstWebhookSubscription.ID,
		Name:        firstWebhookSubscription.Name,
		Description: firstWebhookSubscription.Description,
		WebhookID:   firstWebhookSubscription.WebhookID,
		ObjType:     firstWebhookSubscription.ObjType.String(),
		ObjID:       firstWebhookSubscription.ObjID,
		IsActive:    firstWebhookSubscription.IsActive,
	}
	switch firstWebhookSubscription.ObjType {
	case webhooksubscription.ObjTypeCdnRecord:
		firstCDNRecord, err := global.GvaEntDb.CDNRecord.Query().
			Where(cdnrecord.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstCDNRecord.Domain
		}
	case webhooksubscription.ObjTypeDNSRecord:
		firstDNSRecord, err := global.GvaEntDb.DNSRecord.Query().
			Where(dnsrecord.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstDNSRecord.Domain
		}
	case webhooksubscription.ObjTypeDevice:
		firstDevice, err := global.GvaEntDb.Device.Query().
			Where(device.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstDevice.Name
		}
	}
	successResponse(c, list)
}

// 创建通知订阅
//
//	@Description	创建通知订阅
//	@Summary		创建通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateWebhookSubRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"请求参数"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs [post]
func CreateWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateWebhookSubRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	for _, objID := range requestParams.ObjIDs {
		// 创建通知订阅
		saveWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Create().
			SetName(requestParams.Name).
			SetDescription(requestParams.Description).
			SetWebhookID(requestParams.WebhookID).
			SetObjType(webhooksubscription.ObjType(requestParams.ObjType)).
			SetObjID(objID).
			Save(c)
		if err != nil {
			if global.GvaEnv.DEBUG {
				fmt.Printf("创建通知订阅: %s\n", err.Error())
			}
			errorResponse(c, http.StatusInternalServerError, "创建通知订阅失败: "+err.Error())
			return
		}
		// 记录操作日志
		_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "创建通知订阅")
	}
	successResponse(c, gin.H{
		"ids": requestParams.ObjIDs,
	})
}

// 更新通知订阅
//
//	@Description	更新通知订阅
//	@Summary		更新通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"通知订阅序号"
//	@Param			request	body		types.UpdateWebhookSubRequest					true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id} [patch]
func UpdateWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 更新通知订阅
	saveWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新通知订阅: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新通知订阅失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅")
	successResponse(c, types.GetWebhookSubRequest{
		ID: saveWebhookSubscription.ID,
	})
}

// 更新通知订阅启用状态
//
//	@Description	更新通知订阅启用状态
//	@Summary		更新通知订阅启用状态
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64											true	"通知订阅序号"
//	@Param			request	body		types.UpdateWebhookSubActiveRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								"请求参数"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Failure		500		{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id}/active [patch]
func UpdateWebhookSubActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 更新通知订阅状态
	saveWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.UpdateOne(firstWebhookSubscription).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新通知订阅状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新通知订阅状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅状态")
	successResponse(c, types.GetWebhookSubRequest{
		ID: saveWebhookSubscription.ID,
	})
}

// 删除通知订阅
//
//	@Description	删除通知订阅
//	@Summary		删除通知订阅
//	@Tags			通知订阅
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"通知订阅序号"
//	@Success		200	{object}	APIResponse{data=types.GetWebhookSubRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"请求参数"
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Failure		500	{object}	APIResponse{}									"服务错误"
//	@Router			/webhook-subs/{id} [delete]
func DeleteWebhookSub(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteWebhookSubRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 删除通知订阅
	err = global.GvaEntDb.WebhookSubscription.DeleteOne(firstWebhookSubscription).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除通知订阅: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除通知订阅失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "删除通知订阅")
	successResponse(c, types.GetWebhookSubRequest{
		ID: firstWebhookSubscription.ID,
	})
}
