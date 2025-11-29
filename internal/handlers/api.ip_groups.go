package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/ipgroup"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部IP组
func GetIPGroups(c *gin.Context) {
	ipGroups := make([]types.GetIPGroupsResponse, 0)
	// 查询全部IP组
	allIpGroup, err := global.GvaDb.IpGroup.Query().
		Order(ent.Desc(ipgroup.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	for _, item := range allIpGroup {
		ipGroups = append(ipGroups, types.GetIPGroupsResponse{
			ID:          item.ID,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	successResponse(c, ipGroups)
}

// 获取IP组
func GetIPGroup(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询IP组
	firstIpGroup, err := global.GvaDb.IpGroup.Query().
		Where(ipgroup.ID(cast.ToInt64(id))).
		WithDevices().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 设备列表
	deviceIDs := make([]int64, 0)
	for _, item := range firstIpGroup.Edges.Devices {
		deviceIDs = append(deviceIDs, item.ID)
	}
	successResponse(c, types.GetIPGroupsResponse{
		ID:          firstIpGroup.ID,
		Name:        firstIpGroup.Name,
		Description: firstIpGroup.Description,
		Devices:     deviceIDs,
	})
}

// 创建IP组
func CreateIPGroup(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateIPGroupRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("CreateIPGroup params: %+v\n", requestParams)
	}
	// 创建IP组
	saveIpGroup, err := global.GvaDb.IpGroup.Create().
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "创建IP组")
	successResponse(c, gin.H{
		"id": saveIpGroup.ID,
	})
}

// 更新IP组
func UpdateIPGroup(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	var requestParams types.CreateIPGroupRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("UpdateIPGroup params: %+v\n", requestParams)
	}
	// 查询IP组
	firstIpGroup, err := global.GvaDb.IpGroup.Query().
		Where(ipgroup.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 更新IP组
	saveIpGroup, err := global.GvaDb.IpGroup.UpdateOne(firstIpGroup).
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, saveIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "更新IP组")
	successResponse(c, gin.H{
		"id": saveIpGroup.ID,
	})
}

// 删除IP组
func DeleteIPGroup(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询IP组
	firstIpGroup, err := global.GvaDb.IpGroup.Query().
		Where(ipgroup.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 删除IP组
	err = global.GvaDb.IpGroup.DeleteOne(firstIpGroup).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除IP组: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除IP组失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstIpGroup.ID, operationlog.ObjTypeIPGroup, operationlog.StatusSuccess, "删除IP组")
	successResponse(c, gin.H{
		"id": firstIpGroup.ID,
	})
}

// 获取IP组的设备列表
func GetIPGroupDevices(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询IP组
	firstIpGroup, err := global.GvaDb.IpGroup.Query().
		Where(ipgroup.ID(cast.ToInt64(id))).
		WithDevices().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询IP组失败: "+err.Error())
		return
	}
	// 设备列表
	devices := make([]types.GetDevicesResponse, 0)
	for _, item := range firstIpGroup.Edges.Devices {
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
