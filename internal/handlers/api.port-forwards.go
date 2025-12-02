package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/deviceportforward"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部端口转发
func GetPortForwards(c *gin.Context) {
	portForwards := make([]types.GetPortForwardsResponse, 0)
	// 查询全部端口转发
	allPortForward, err := global.GvaDb.DevicePortForward.Query().
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
			DeviceID:    item.Edges.Device.ID,
			DeviceName:  item.Edges.Device.Name,
		})
	}
	successResponse(c, portForwards)
}

// 获取端口转发
func GetPortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.GetPortForwardRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询端口转发
	firstPortForward, err := global.GvaDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		WithDevice().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	successResponse(c, types.GetPortForwardsResponse{
		ID:          firstPortForward.ID,
		Name:        firstPortForward.Name,
		Description: firstPortForward.Description,
		LocalPort:   firstPortForward.LocalPort,
		RemoteHost:  firstPortForward.RemoteHost,
		RemotePort:  firstPortForward.RemotePort,
		Protocols:   firstPortForward.Protocols,
		AutoForward: firstPortForward.AutoForward,
		IsActive:    firstPortForward.IsActive,
		DeviceID:    firstPortForward.Edges.Device.ID,
		DeviceName:  firstPortForward.Edges.Device.Name,
	})
}

// 创建端口转发
func CreatePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.CreatePortForwardRequest
	if err := gin_param.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建端口转发
	savePortForward, err := global.GvaDb.DevicePortForward.Create().
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		SetLocalPort(requestParams.LocalPort).
		SetRemoteHost(requestParams.RemoteHost).
		SetRemotePort(requestParams.RemotePort).
		SetProtocols(requestParams.Protocols).
		SetDeviceID(requestParams.DeviceID).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建端口转发: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建端口转发失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "创建端口转发")
	successResponse(c, gin.H{
		"id": savePortForward.ID,
	})
}

// 更新端口转发
func UpdatePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询端口转发
	firstPortForward, err := global.GvaDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	// 更新端口转发
	savePortForward, err := global.GvaDb.DevicePortForward.UpdateOne(firstPortForward).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		SetLocalPort(requestParams.LocalPort).
		SetRemoteHost(requestParams.RemoteHost).
		SetRemotePort(requestParams.RemotePort).
		SetProtocols(requestParams.Protocols).
		SetDeviceID(requestParams.DeviceID).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新端口转发: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新端口转发失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发")
	successResponse(c, gin.H{
		"id": savePortForward.ID,
	})
}

// 更新端口转发防火墙状态
func UpdatePortForwardForward(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardForwardRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询端口转发
	firstPortForward, err := global.GvaDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	// 更新端口转发状态
	savePortForward, err := global.GvaDb.DevicePortForward.UpdateOne(firstPortForward).
		SetAutoForward(requestParams.AutoForward).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新端口转发防火墙状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新端口转发防火墙状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发防火墙状态")
	successResponse(c, gin.H{
		"id": savePortForward.ID,
	})
}

// 更新端口转发启用状态
func UpdatePortForwardActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdatePortForwardActiveRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询端口转发
	firstPortForward, err := global.GvaDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	// 更新端口转发状态
	savePortForward, err := global.GvaDb.DevicePortForward.UpdateOne(firstPortForward).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新端口转发启用状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新端口转发启用状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, savePortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "更新端口转发启用状态")
	successResponse(c, gin.H{
		"id": savePortForward.ID,
	})
}

// 删除端口转发
func DeletePortForward(c *gin.Context) {
	// 请求参数
	var requestParams types.DeletePortForwardRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询端口转发
	firstPortForward, err := global.GvaDb.DevicePortForward.Query().
		Where(deviceportforward.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询端口转发失败: "+err.Error())
		return
	}
	// 删除端口转发
	err = global.GvaDb.DevicePortForward.DeleteOne(firstPortForward).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除端口转发: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除端口转发失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstPortForward.ID, operationlog.ObjTypeDevicePortForward, operationlog.StatusSuccess, "删除端口转发")
	successResponse(c, gin.H{
		"id": firstPortForward.ID,
	})
}
