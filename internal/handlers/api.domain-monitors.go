package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/domainmonitor"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部域名监控
func GetDomainMonitors(c *gin.Context) {
	domainMonitors := make([]types.GetDomainMonitorResponse, 0)
	// 查询全部域名监控
	allDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Order(ent.Desc(domainmonitor.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	for _, item := range allDomainMonitor {
		list := types.GetDomainMonitorResponse{
			ID:            item.ID,
			Name:          item.Name,
			Domain:        item.Domain,
			CheckInterval: item.CheckInterval,
			CheckProtocol: item.CheckProtocol.String(),
			IsActive:      item.IsActive,
			CurrentStatus: item.CurrentStatus.String(),
			LastSuccessAt: item.LastSuccessAt,
			LastFailureAt: item.LastFailureAt,
		}
		// 查询通知订阅
		allWebhookSub, _ := global.GvaEntDb.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			All(c)
		for _, item := range allWebhookSub {
			list.Webhooks = append(list.Webhooks, item.ID)
		}
		domainMonitors = append(domainMonitors, list)
	}
	successResponse(c, domainMonitors)
}

// 获取域名监控
func GetDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	response := types.GetDomainMonitorResponse{
		ID:            firstDomainMonitor.ID,
		Name:          firstDomainMonitor.Name,
		Domain:        firstDomainMonitor.Domain,
		CheckInterval: firstDomainMonitor.CheckInterval,
		CheckProtocol: firstDomainMonitor.CheckProtocol.String(),
		IsActive:      firstDomainMonitor.IsActive,
		CurrentStatus: firstDomainMonitor.CurrentStatus.String(),
		LastSuccessAt: firstDomainMonitor.LastSuccessAt,
		LastFailureAt: firstDomainMonitor.LastFailureAt,
	}
	// 查询通知订阅
	allWebhookSub, _ := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeDomainMonitor)).
		Where(webhooksubscription.ObjIDEQ(requestParams.ID)).
		All(c)
	for _, item := range allWebhookSub {
		response.Webhooks = append(response.Webhooks, item.ID)
	}
	successResponse(c, response)
}

// 创建域名监控
func CreateDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDomainMonitorRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建域名监控
	saveDomainMonitor, err := global.GvaEntDb.DomainMonitor.Create().
		SetName(requestParams.Name).
		SetDomain(requestParams.Domain).
		SetCheckInterval(requestParams.CheckInterval).
		SetCheckProtocol(domainmonitor.CheckProtocol(requestParams.CheckProtocol)).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建域名监控: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建域名监控失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "创建域名监控")
	successResponse(c, gin.H{
		"id": saveDomainMonitor.ID,
	})
}

// 更新域名监控
func UpdateDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	// 更新域名监控
	saveDomainMonitor, err := global.GvaEntDb.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetName(requestParams.Name).
		SetCheckInterval(requestParams.CheckInterval).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新域名监控: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新域名监控失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "更新域名监控")
	successResponse(c, gin.H{
		"id": saveDomainMonitor.ID,
	})
}

// 更新域名监控启用状态
func UpdateDomainMonitorActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDomainMonitorActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	// 更新域名监控状态
	saveDomainMonitor, err := global.GvaEntDb.DomainMonitor.UpdateOne(firstDomainMonitor).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新域名监控启用状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新域名监控启用状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "更新域名监控启用状态")
	successResponse(c, gin.H{
		"id": saveDomainMonitor.ID,
	})
}

// 删除域名监控
func DeleteDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}
	// 删除域名监控
	err = global.GvaEntDb.DomainMonitor.DeleteOne(firstDomainMonitor).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除域名监控: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除域名监控失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstDomainMonitor.ID, operationlog.ObjTypeDomainMonitor, operationlog.StatusSuccess, "删除域名监控")
	successResponse(c, gin.H{
		"id": firstDomainMonitor.ID,
	})
}

// 检查域名监控
func CheckDomainMonitor(c *gin.Context) {
	// 请求参数
	var requestParams types.CheckDomainMonitorRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询域名监控
	firstDomainMonitor, err := global.GvaEntDb.DomainMonitor.Query().
		Where(domainmonitor.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询域名监控失败: "+err.Error())
		return
	}

	// 域名监控任务
	domainMonitor := backend.NewDomainMonitorService(global.GvaEntDb, global.GvaGoCron)

	// 执行单个监控任务
	err = domainMonitor.ExecuteMonitorTask(c, slog.With(
		slog.Int64("id", firstDomainMonitor.ID),
		slog.String("domain", firstDomainMonitor.Domain),
		slog.String("protocol", firstDomainMonitor.CheckProtocol.String()),
	), firstDomainMonitor)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "检查域名监控失败: "+err.Error())
		return
	}
	successResponse(c, gin.H{
		"id": firstDomainMonitor.ID,
	})
}
