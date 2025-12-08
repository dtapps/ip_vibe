package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/ipgroup"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/webhook_sub_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"
)

// 获取全部IP组
func GetIPGroups(c *gin.Context) {
	ipGroups := make([]types.GetIPGroupResponse, 0)
	// 查询全部IP组
	allIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Order(ent.Desc(ipgroup.FieldCreatedAt)).
		WithDevices().
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	for _, item := range allIpGroup {
		list := types.GetIPGroupResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		}
		// 设备列表
		for _, item := range item.Edges.Devices {
			list.Devices = append(list.Devices, item.ID)
		}
		// 查询通知订阅
		allWebhookSub, _ := global.GvaEntDb.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeIPGroup)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			All(c)
		for _, item := range allWebhookSub {
			list.Webhooks = append(list.Webhooks, item.ID)
		}
		ipGroups = append(ipGroups, list)
	}
	successResponse(c, ipGroups)
}

// 获取IP组
func GetIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.GetIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询IP组
	firstIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Where(ipgroup.IDEQ(requestParams.ID)).
		WithDevices().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 通知订阅
	response := types.GetIPGroupResponse{
		ID:          firstIpGroup.ID,
		Name:        firstIpGroup.Name,
		Description: firstIpGroup.Description,
	}
	// 设备列表
	for _, item := range firstIpGroup.Edges.Devices {
		response.Devices = append(response.Devices, item.ID)
	}
	// 查询通知订阅
	allWebhookSub, _ := global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeIPGroup)).
		Where(webhooksubscription.ObjIDEQ(requestParams.ID)).
		All(c)
	for _, item := range allWebhookSub {
		response.Webhooks = append(response.Webhooks, item.ID)
	}
	successResponse(c, response)
}

// 创建IP组
func CreateIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateIPGroupRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建IP组
	saveIpGroup, err := global.GvaEntDb.IpGroup.Create().
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		AddDeviceIDs(requestParams.Devices...).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建IP组: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建IP组失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "创建IP组")

	// 创建通知订阅
	_ = webhook_sub_helper.Create(c, requestParams.WebhookIDs, webhooksubscription.ObjTypeIPGroup, saveIpGroup.ID)

	successResponse(c, gin.H{
		"id": saveIpGroup.ID,
	})
}

// 更新IP组
func UpdateIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询IP组
	firstIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Where(ipgroup.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 更新IP组
	saveIpGroup, err := global.GvaEntDb.IpGroup.UpdateOne(firstIpGroup).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		ClearDevices().
		AddDeviceIDs(requestParams.Devices...).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新IP组: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新IP组失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "更新IP组")

	// 更新通知订阅
	_ = webhook_sub_helper.Update(c, requestParams.WebhookIDs, webhooksubscription.ObjTypeIPGroup, saveIpGroup.ID)

	successResponse(c, gin.H{
		"id": saveIpGroup.ID,
	})
}

// 删除IP组
func DeleteIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询IP组
	firstIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Where(ipgroup.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 删除IP组
	err = global.GvaEntDb.IpGroup.DeleteOne(firstIpGroup).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除IP组: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除IP组失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "删除IP组")

	// 删除通知订阅
	_ = webhook_sub_helper.Delete(c, webhooksubscription.ObjTypeIPGroup, firstIpGroup.ID)

	successResponse(c, gin.H{
		"id": firstIpGroup.ID,
	})
}

// 获取IP组的设备列表
func GetIPGroupDevices(c *gin.Context) {
	// 请求参数
	var requestParams types.GetIPGroupDevicesRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询IP组
	firstIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Where(ipgroup.IDEQ(requestParams.ID)).
		WithDevices().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 设备列表
	devices := make([]types.GetDeviceResponse, 0)
	for _, item := range firstIpGroup.Edges.Devices {
		devices = append(devices, types.GetDeviceResponse{
			ID:              item.ID,
			Code:            item.Code,
			Name:            item.Name,
			Description:     item.Description,
			IPV4:            item.IPV4,
			IPV6:            item.IPV6,
			LastReportedAt:  item.LastReportedAt,
			LastHeartbeatAt: item.LastHeartbeatAt,
		})
	}
	successResponse(c, devices)
}

// 检查IP组健康状态
func HealthcheckIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthcheckIPGroupRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询IP组
	firstIpGroup, err := global.GvaEntDb.IpGroup.Query().
		Where(ipgroup.IDEQ(requestParams.ID)).
		WithDevices().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}

	// 检查健康状态
	results := make([][]backend.ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(c)
	eg.SetLimit(10) // 最多 10 个并发

	for _, item := range firstIpGroup.Edges.Devices {
		item := item // 创建副本以避免闭包问题
		if item.IPV4 != "" {
			// 检测 IP 是否可通过 HTTP/HTTPS 访问
			eg.Go(func() error {
				healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV4, 80, 443)
				results = append(results, healthCheckResult.ToPairsForIP())
				return nil
			})
		}
		if item.IPV6 != "" {
			// 检测 IP 是否可通过 HTTP/HTTPS 访问
			eg.Go(func() error {
				healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV6, 80, 443)
				results = append(results, healthCheckResult.ToPairsForIP())
				return nil
			})
		}
	}

	if err = eg.Wait(); err != nil {
		errorResponse(c, http.StatusInternalServerError, "IP组健康状态检查失败: "+err.Error())
		return
	}

	successResponse(c, results)
}
