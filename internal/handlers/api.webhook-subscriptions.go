package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/device"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部通知订阅
func GetWebhookSubscriptions(c *gin.Context) {
	webhooks := make([]types.GetWebhookSubscriptionsResponse, 0)
	// 查询全部通知订阅
	allWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		Order(ent.Desc(webhooksubscription.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	for _, item := range allWebhookSubscription {
		list := types.GetWebhookSubscriptionsResponse{
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
			firstCDNRecord, err := global.GvaDb.CDNRecord.Query().
				Where(cdnrecord.IDEQ(item.ObjID)).
				First(c)
			if err == nil {
				list.ObjName = firstCDNRecord.Domain
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
		webhooks = append(webhooks)
	}
	successResponse(c, webhooks)
}

// 获取通知订阅
func GetWebhookSubscription(c *gin.Context) {
	// 请求参数
	var requestParams types.GetWebhookSubscriptionRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	list := types.GetWebhookSubscriptionsResponse{
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
		firstCDNRecord, err := global.GvaDb.CDNRecord.Query().
			Where(cdnrecord.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstCDNRecord.Domain
		}
	case webhooksubscription.ObjTypeDNSRecord:
		firstDNSRecord, err := global.GvaDb.DNSRecord.Query().
			Where(dnsrecord.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstDNSRecord.Domain
		}
	case webhooksubscription.ObjTypeDevice:
		firstDevice, err := global.GvaDb.Device.Query().
			Where(device.IDEQ(firstWebhookSubscription.ObjID)).
			First(c)
		if err == nil {
			list.ObjName = firstDevice.Name
		}
	}
	successResponse(c, list)
}

// 创建通知订阅
func CreateWebhookSubscription(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateWebhookSubscriptionRequest
	if err := gin_param.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	for _, objID := range requestParams.ObjIDs {
		// 创建通知订阅
		saveWebhookSubscription, err := global.GvaDb.WebhookSubscription.Create().
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
		_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "创建通知订阅")
	}
	successResponse(c, gin.H{
		"ids": requestParams.ObjIDs,
	})
}

// 更新通知订阅
func UpdateWebhookSubscription(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubscriptionRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 更新通知订阅
	saveWebhookSubscription, err := global.GvaDb.WebhookSubscription.UpdateOne(firstWebhookSubscription).
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅")
	successResponse(c, gin.H{
		"id": saveWebhookSubscription.ID,
	})
}

// 更新通知订阅状态
func UpdateWebhookSubscriptionActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateWebhookSubscriptionActiveRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 更新通知订阅状态
	saveWebhookSubscription, err := global.GvaDb.WebhookSubscription.UpdateOne(firstWebhookSubscription).
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, saveWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "更新通知订阅状态")
	successResponse(c, gin.H{
		"id": saveWebhookSubscription.ID,
	})
}

// 删除通知订阅
func DeleteWebhookSubscription(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteWebhookSubscriptionRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询通知订阅
	firstWebhookSubscription, err := global.GvaDb.WebhookSubscription.Query().
		Where(webhooksubscription.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询通知订阅失败: "+err.Error())
		return
	}
	// 删除通知订阅
	err = global.GvaDb.WebhookSubscription.DeleteOne(firstWebhookSubscription).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除通知订阅: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除通知订阅失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstWebhookSubscription.ID, operationlog.ObjTypeWebhookSubscription, operationlog.StatusSuccess, "删除通知订阅")
	successResponse(c, gin.H{
		"id": firstWebhookSubscription.ID,
	})
}
