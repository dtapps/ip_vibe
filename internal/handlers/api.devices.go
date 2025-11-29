package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/ent/deviceipreport"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/webhooksubscription"
	"vibe_ip/internal/global"
	"vibe_ip/internal/services"
	"vibe_ip/internal/types"
	"vibe_ip/internal/utils"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部设备
func GetDevices(c *gin.Context) {
	devices := make([]types.GetDevicesResponse, 0)
	// 查询全部设备
	allDevice, err := global.GvaDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	for _, item := range allDevice {
		devices = append(devices, types.GetDevicesResponse{
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

// 获取设备
func GetDevice(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	successResponse(c, types.GetDevicesResponse{
		ID:              firstDevice.ID,
		Code:            firstDevice.Code,
		Name:            firstDevice.Name,
		Description:     firstDevice.Description,
		IPV4:            firstDevice.IPV4,
		IPV6:            firstDevice.IPV6,
		LastReportedAt:  firstDevice.LastReportedAt,
		LastHeartbeatAt: firstDevice.LastHeartbeatAt,
	})
}

// 创建设备
func CreateDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDeviceRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("CreateDevice params: %+v\n", requestParams)
	}
	// 创建设备
	saveDevice, err := global.GvaDb.Device.Create().
		SetCode(requestParams.Code).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建设备失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "创建设备")
	successResponse(c, gin.H{
		"id": saveDevice.ID,
	})
}

// 更新设备
func UpdateDevice(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	var requestParams types.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("UpdateDevice params: %+v\n", requestParams)
	}
	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	// 更新设备
	saveDevice, err := global.GvaDb.Device.UpdateOne(firstDevice).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新设备失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "更新设备")
	successResponse(c, gin.H{
		"id": saveDevice.ID,
	})
}

// 删除设备
func DeleteDevice(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	// 获取自身设备ID
	selfDeviceCode, err := machineid.ID()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取设备ID失败: "+err.Error())
		return
	}
	// 检查是否是自身设备
	if firstDevice.Code == selfDeviceCode {
		errorResponse(c, http.StatusForbidden, "不能删除自身设备")
		return
	}
	// 删除设备
	err = global.GvaDb.Device.DeleteOne(firstDevice).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除设备: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除设备失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "删除设备")
	successResponse(c, gin.H{
		"id": firstDevice.ID,
	})
}

// 获取设备的IP历史记录
func GetDeviceIpHistory(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	// 查询设备IP历史记录
	firstDeviceReports, err := firstDevice.QueryIPReports().
		Order(ent.Desc(deviceipreport.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备IP历史记录失败: "+err.Error())
		return
	}
	// 设备列表
	ips := make([]types.GetDeviceIpHistoryRequest, 0)
	for _, item := range firstDeviceReports {
		ips = append(ips, types.GetDeviceIpHistoryRequest{
			IPV4:      item.IPV4,
			IPV6:      item.IPV6,
			CreatedAt: item.CreatedAt,
		})
	}
	successResponse(c, ips)
}

// 上报设备IP
func ReportDeviceIp(c *gin.Context) {
	// 请求参数
	var requestParams types.ReportDeviceIpRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("ReportDeviceIp params: %+v\n", requestParams)
	}
	// 判断IPV4和IPV6是否为空
	if requestParams.IPV4Addr == "" && requestParams.IPV6Addr == "" {
		errorResponse(c, http.StatusBadRequest, "IPV4和IPV6不能同时为空")
		return
	}

	// 获取当前设备ID
	selfDeviceCode, err := machineid.ID()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "获取当前设备ID失败: "+err.Error())
		return
	}
	if requestParams.DeviceCode == selfDeviceCode {
		errorResponse(c, http.StatusForbidden, "不能上报自身IP")
		return
	}

	// 是否通知
	isNotify := false
	notifyData := utils.TemplateData{
		EventType:  "IP_UPDATE",
		EventTime:  time.Now().Format("2006-01-02 15:04:05"),
		EventBody:  "IP地址更新事件",
		DeviceCode: requestParams.DeviceCode,
		DeviceName: requestParams.DeviceName,
		IPv4Addr:   requestParams.IPV4Addr,
		IPv6Addr:   requestParams.IPV6Addr,
	}

	// 描述
	msg := ""

	// 开始事务
	tx, err := global.GvaDb.Tx(c)
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

		// 如果设备不存在和设备名称为空，则使用设备编号作为设备名称
		if requestParams.DeviceName == "" {
			requestParams.DeviceName = requestParams.DeviceCode
		}

		// 创建设备
		saveDevice, err := tx.Device.Create().
			SetCode(requestParams.DeviceCode).
			SetName(requestParams.DeviceName).
			SetDescription(requestParams.Description).
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
		requestParams.ID = saveDevice.ID
		// 记录操作日志
		_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(c, saveDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "接口创建设备")

		// 创建设备IP报告
		saveDeviceIPReport, err := tx.DeviceIpReport.Create().
			SetDevice(saveDevice).
			SetIPV4(requestParams.IPV4Addr).
			SetIPV6(requestParams.IPV6Addr).
			Save(c)
		if err != nil {
			_ = tx.Rollback()
			if global.GvaEnv.DEBUG {
				fmt.Printf("创建设备IP报告: %s\n", err.Error())
			}
			errorResponse(c, http.StatusInternalServerError, "创建设备IP报告失败: "+err.Error())
			return
		}
		// 记录操作日志
		_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(c, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "接口创建设备IP报告")

		isNotify = true

		msg = "创建设备IP成功"
	} else {
		requestParams.ID = firstDevice.ID
		requestParams.DeviceName = firstDevice.Name
		if firstDevice.IPV4 != requestParams.IPV4Addr || firstDevice.IPV6 != requestParams.IPV6Addr {
			// 更新设备IP
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetIPV4(requestParams.IPV4Addr).
				SetIPV6(requestParams.IPV6Addr).
				SetLastReportedAt(time.Now()).
				SetLastHeartbeatAt(time.Now())
			if requestParams.DeviceName != "" && firstDevice.Name != requestParams.DeviceName {
				updateDevice.SetName(requestParams.DeviceName)
			}
			if requestParams.Description != "" && firstDevice.Description != requestParams.Description {
				updateDevice.SetDescription(requestParams.Description)
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
			// 记录操作日志
			_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataUpdate(c, firstDevice.ID, operationlog.ObjTypeDevice, operationlog.StatusSuccess, "接口更新设备")

			// 创建设备IP报告
			saveDeviceIPReport, err := tx.DeviceIpReport.Create().
				SetDevice(firstDevice).
				SetIPV4(requestParams.IPV4Addr).
				SetIPV6(requestParams.IPV6Addr).
				Save(c)
			if err != nil {
				_ = tx.Rollback()
				if global.GvaEnv.DEBUG {
					fmt.Printf("创建设备IP报告: %s\n", err.Error())
				}
				errorResponse(c, http.StatusInternalServerError, "创建设备IP报告失败: "+err.Error())
				return
			}
			// 记录操作日志
			_ = global.GvaOperationLogHelper.SetTx(tx).RecordDataCreate(c, saveDeviceIPReport.ID, operationlog.ObjTypeDeviceIPReport, operationlog.StatusSuccess, "接口创建设备IP报告")

			isNotify = true

			msg = "更新设备IP成功"
		} else {
			// 更新设备心跳时间
			updateDevice := tx.Device.UpdateOne(firstDevice).
				SetLastHeartbeatAt(time.Now())
			if requestParams.DeviceName != "" && firstDevice.Name != requestParams.DeviceName {
				updateDevice.SetName(requestParams.DeviceName)
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
	_ = tx.Commit()

	// 发送通知
	if isNotify {
		go func() {
			_ = services.NewWebhookService(global.GvaDb).Send(context.Background(), webhooksubscription.ObjTypeDevice, requestParams.ID, notifyData)
		}()
	}

	successResponse(c, gin.H{
		"id":  requestParams.ID,
		"msg": msg,
	})
}

func GetDevicePortForwards(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	portForwards := make([]types.GetPortForwardsResponse, 0)
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
		portForwards = append(portForwards, types.GetPortForwardsResponse{
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
