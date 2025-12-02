package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/operationlog"
	"vibe_ip/ent/platformconfig"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_param"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部配置
func GetConfigs(c *gin.Context) {
	configs := make([]types.GetConfigsResponse, 0)
	// 查询全部配置
	allConfig, err := global.GvaDb.PlatformConfig.Query().
		Order(ent.Desc(platformconfig.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询配置失败: "+err.Error())
		return
	}
	for _, item := range allConfig {
		configs = append(configs, types.GetConfigsResponse{
			ID:          item.ID,
			Capability:  item.Capability.String(),
			Name:        item.Name,
			Description: item.Description,
			Platform:    item.Platform.String(),
			ConfigJSON:  item.ConfigJSON,
		})
	}
	successResponse(c, configs)
}

// 获取配置
func GetConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.GetConfigRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询配置
	firstConfig, err := global.GvaDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询配置失败: "+err.Error())
		return
	}
	successResponse(c, types.GetConfigsResponse{
		ID:          firstConfig.ID,
		Capability:  firstConfig.Capability.String(),
		Name:        firstConfig.Name,
		Description: firstConfig.Description,
		Platform:    firstConfig.Platform.String(),
		ConfigJSON:  firstConfig.ConfigJSON,
	})
}

// 创建配置
func CreateConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateConfigRequest
	if err := gin_param.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建配置
	saveConfig, err := global.GvaDb.PlatformConfig.Create().
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		SetPlatform(platformconfig.Platform(requestParams.Platform)).
		SetConfigJSON([]byte(requestParams.ConfigJSON)).
		SetCapability(platformconfig.Capability(requestParams.Capability)).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建配置: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建配置失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "创建配置")
	successResponse(c, gin.H{
		"id": saveConfig.ID,
	})
}

// 更新配置
func UpdateConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateConfigRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询配置
	firstConfig, err := global.GvaDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询配置失败: "+err.Error())
		return
	}
	// 更新配置
	saveConfig, err := global.GvaDb.PlatformConfig.UpdateOne(firstConfig).
		SetName(requestParams.Name).
		SetDescription(requestParams.Description).
		SetConfigJSON([]byte(requestParams.ConfigJSON)).
		SetCapability(platformconfig.Capability(requestParams.Capability)).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新配置: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新配置失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "更新配置")
	successResponse(c, gin.H{
		"id": saveConfig.ID,
	})
}

// 删除配置
func DeleteConfig(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteConfigRequest
	if err := gin_param.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询配置
	firstConfig, err := global.GvaDb.PlatformConfig.Query().
		Where(platformconfig.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询配置失败: "+err.Error())
		return
	}
	// 删除配置
	err = global.GvaDb.PlatformConfig.DeleteOne(firstConfig).Exec(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("删除配置: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "删除配置失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstConfig.ID, operationlog.ObjTypePlatformConfig, operationlog.StatusSuccess, "删除配置")
	successResponse(c, gin.H{
		"id": firstConfig.ID,
	})
}
