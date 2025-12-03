package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/device"
	"vibe_ip/ent/deviceipreport"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// 获取全部设备
func GetDevices(c *gin.Context) {
	devices := make([]types.GetDeviceResponse, 0)
	// 查询全部设备
	allDevice, err := global.GvaDb.Device.Query().
		Order(ent.Desc(device.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	for _, item := range allDevice {
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

// 获取设备
func GetDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.GetDeviceRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}
	successResponse(c, types.GetDeviceResponse{
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
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
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
	var requestParams types.UpdateDeviceRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
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
	var requestParams types.DeleteDeviceRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
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
	var requestParams types.GetDeviceIpHistoryRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
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
	ips := make([]types.GetDeviceIpHistoryResponse, 0)
	for _, item := range firstDeviceReports {
		ips = append(ips, types.GetDeviceIpHistoryResponse{
			IPV4:      item.IPV4,
			IPV6:      item.IPV6,
			CreatedAt: item.CreatedAt,
		})
	}
	successResponse(c, ips)
}

// 检查设备健康状态
func HealthcheckDevice(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthcheckDeviceRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询设备
	firstDevice, err := global.GvaDb.Device.Query().
		Where(device.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询设备失败: "+err.Error())
		return
	}

	// 检查健康状态
	results := make([][]backend.ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(c)
	eg.SetLimit(10) // 最多 10 个并发

	if firstDevice.IPV4 != "" {
		// 检测 IP 是否可通过 HTTP/HTTPS 访问
		eg.Go(func() error {
			healthCheckResult, _ := backend.CheckIPHealth(egCtx, firstDevice.IPV4, 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}
	if firstDevice.IPV6 != "" {
		// 检测 IP 是否可通过 HTTP/HTTPS 访问
		eg.Go(func() error {
			healthCheckResult, _ := backend.CheckIPHealth(egCtx, firstDevice.IPV6, 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForIP())
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		errorResponse(c, http.StatusInternalServerError, "设备健康状态检查失败: "+err.Error())
		return
	}

	successResponse(c, results)
}
