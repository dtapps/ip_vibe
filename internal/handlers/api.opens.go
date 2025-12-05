package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/operation_logs_helper"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/gin-gonic/gin"
)

// 设备上报IP
func DeviceReportDeviceIp(c *gin.Context) {
	// 请求参数
	var requestParams types.DeviceReportDeviceIpRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 判断IPV4和IPV6是否为空
	if requestParams.IPV4Addr == "" && requestParams.IPV6Addr == "" {
		errorResponse(c, http.StatusBadRequest, "IPV4和IPV6不能同时为空")
		return
	}

	// 异步处理所需数据
	isNotify := false // 是否需要发送通知和记录IP报告
	var deviceID int64
	var isNewDevice bool // 记录是否为新设备，用于异步日志记录
	var isIPChanged bool // 记录IP是否发生变化，用于异步日志记录

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventType:  "IP_UPDATE",
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		EventBody:  "IP地址更新事件",
		DeviceCode: requestParams.DeviceCode,
		DeviceName: requestParams.DeviceName,
		IPv4Addr:   requestParams.IPV4Addr,
		IPv6Addr:   requestParams.IPV6Addr,
	}

	msg := ""

	// 开始事务
	tx, err := global.GvaEntDb.Tx(c)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "开始事务失败: "+err.Error())
		return
	}

	// 检查设备是否存在
	firstDevice, err := tx.Device.Query().
		Where(device.CodeEQ(requestParams.DeviceCode)).
		First(c)
	if err != nil {
		if !ent.IsNotFound(err) {
			_ = tx.Rollback()
			errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
			return
		}

		// 如果设备不存在：创建设备
		isNewDevice = true
		if requestParams.DeviceName == "" {
			requestParams.DeviceName = requestParams.DeviceCode
		}

		saveDevice, err := tx.Device.Create().
			SetCode(requestParams.DeviceCode).
			SetName(requestParams.DeviceName).
			SetIPV4(requestParams.IPV4Addr).
			SetIPV6(requestParams.IPV6Addr).
			SetLastReportedAt(time.Now()).
			SetLastHeartbeatAt(time.Now()).
			Save(c)
		if err != nil {
			_ = tx.Rollback()
			if global.GvaEnv.DEBUG {
				fmt.Printf("创建设备: %s\n", err.Error())
			}
			errorResponse(c, http.StatusInternalServerError, "创建设备失败: "+err.Error())
			return
		}

		deviceID = saveDevice.ID
		isNotify = true // 新设备必须通知和记录报告
		isIPChanged = true
		msg = "创建设备IP成功"

	} else {
		// 设备存在：检查 IP 是否变化
		deviceID = firstDevice.ID
		requestParams.DeviceName = firstDevice.Name // 使用数据库中的设备名进行通知

		if firstDevice.IPV4 != requestParams.IPV4Addr || firstDevice.IPV6 != requestParams.IPV6Addr {
			// IP 发生变化
			isIPChanged = true
			isNotify = true

			// 更新设备IP和心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetIPV4(requestParams.IPV4Addr).
				SetIPV6(requestParams.IPV6Addr).
				SetLastReportedAt(time.Now()).
				SetLastHeartbeatAt(time.Now())

			// 如果设备名有更新
			if requestParams.DeviceName != "" && firstDevice.Name != requestParams.DeviceName {
				updateDevice.SetName(requestParams.DeviceName)
				// 确保异步通知使用新的设备名
				notifyData.DeviceName = requestParams.DeviceName
			}

			err = updateDevice.Exec(c)
			if err != nil {
				_ = tx.Rollback()
				if global.GvaEnv.DEBUG {
					fmt.Printf("更新设备IP: %s\n", err.Error())
				}
				errorResponse(c, http.StatusInternalServerError, "更新设备IP失败: "+err.Error())
				return
			}

			msg = "更新设备IP成功"

		} else {
			// IP 未改变：仅更新心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetLastHeartbeatAt(time.Now())

			// 如果设备名有更新
			if requestParams.DeviceName != "" && firstDevice.Name != requestParams.DeviceName {
				updateDevice.SetName(requestParams.DeviceName)
				notifyData.DeviceName = requestParams.DeviceName
			}

			err = updateDevice.Exec(c)
			if err != nil {
				_ = tx.Rollback()
				if global.GvaEnv.DEBUG {
					fmt.Printf("更新设备心跳时间: %s\n", err.Error())
				}
				errorResponse(c, http.StatusInternalServerError, "更新设备心跳时间失败: "+err.Error())
				return
			}
			msg = "设备IP未改变，不添加IP报告"
		}
	}

	// 提交核心事务：立即释放独占锁，这是解决卡死的关键！
	if err := tx.Commit(); err != nil {
		errorResponse(c, http.StatusInternalServerError, "提交事务失败: "+err.Error())
		return
	}

	// --- 异步处理非核心写入和通知 ---
	// 使用独立的 Go Routine 和 Background Context，不阻塞 HTTP 响应
	go func(asyncCtx context.Context, asyncDb *ent.Client, asyncLogHelper *operation_logs_helper.OperationLogHelper) {

		// 记录 Device 相关的操作日志
		// 只在创建设备或IP更新时记录设备操作日志，避免高频心跳日志
		if isNewDevice {
			_ = asyncLogHelper.RecordDataCreate(asyncCtx, deviceID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "接口创建设备")
		} else if isIPChanged {
			_ = asyncLogHelper.RecordDataUpdate(asyncCtx, deviceID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "接口更新设备")
		}

		// 记录 IP 报告并发送通知
		if isNotify {
			// 创建设备IP报告
			saveDeviceIPReport, err := asyncDb.DeviceIpReport.Create().
				SetDeviceID(deviceID).
				SetIPV4(requestParams.IPV4Addr).
				SetIPV6(requestParams.IPV6Addr).
				Save(asyncCtx)
			if err != nil {
				// 异步写入失败，仅打印日志
				fmt.Printf("[异步写入失败] 创建设备IP报告: %s\n", err.Error())
			} else {
				// 记录设备IP报告的操作日志
				_ = asyncLogHelper.RecordDataCreate(asyncCtx, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "接口创建设备IP报告")
			}

			// 发送通知 (你原来就在 go func 中，现在并入这个更集中的异步块)
			_ = backend.NewWebhookService(asyncDb).Send(asyncCtx, webhooksubscription.ObjTypeDevice, deviceID, notifyData)
		}
	}(context.Background(), global.GvaEntDb, global.GvaOperationLogHelper)

	successResponse(c, gin.H{
		"id":  deviceID,
		"msg": msg,
	})
}

// 获取设备端口转发
func DevicePortForwards(c *gin.Context) {
	// 请求参数
	var requestParams types.DevicePortForwardsRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaEntDb.Device.Query().
		Where(device.CodeEQ(requestParams.DeviceCode)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	portForwards := make([]types.GetPortForwardResponse, 0)
	// 查询端口转发
	allPortForward, err := firstDevice.QueryPortForwards().
		Where(deviceportforward.IsActiveEQ(true)).
		Order(ent.Desc(deviceportforward.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	for _, item := range allPortForward {
		portForwards = append(portForwards, types.GetPortForwardResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			LocalPort:   item.LocalPort,
			RemoteHost:  item.RemoteHost,
			RemotePort:  item.RemotePort,
			Protocols:   item.Protocols,
			AutoForward: item.AutoForward,
			IsActive:    item.IsActive,
		})
	}
	successResponse(c, portForwards)
}
