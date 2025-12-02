package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取全部CDN记录
func GetCdnRecords(c *gin.Context) {
	records := make([]types.GetCdnRecordsResponse, 0)
	// 查询全部CDN记录
	allCdnRecord, err := global.GvaDb.CDNRecord.Query().
		WithConfig().
		WithIPGroup().
		Order(ent.Desc(cdnrecord.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询记录失败: "+err.Error())
		return
	}
	for _, item := range allCdnRecord {
		list := types.GetCdnRecordsResponse{
			ID:             item.ID,
			ConfigID:       item.ConfigID,
			Platform:       item.Edges.Config.Platform.String(),
			RecordID:       item.RecordID,
			Domain:         item.Domain,
			Type:           item.Type.String(),
			OriginProtocol: item.OriginProtocol.String(),
			HTTPPort:       item.HTTPPort,
			HTTPSPort:      item.HTTPSPort,
			IsActive:       item.IsActive,
			LastRefreshAt:  item.LastRefreshAt,
		}
		if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
			list.IPGroupID = ipGroup.ID
			list.IPGroupName = ipGroup.Name
		}
		records = append(records, list)
	}
	successResponse(c, records)
}

// 获取CDN记录
func GetCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		WithConfig().
		WithIPGroup().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	response := types.GetCdnRecordsResponse{
		ID:             firstCdnRecord.ID,
		ConfigID:       firstCdnRecord.ConfigID,
		Platform:       firstCdnRecord.Edges.Config.Platform.String(),
		RecordID:       firstCdnRecord.RecordID,
		Domain:         firstCdnRecord.Domain,
		Type:           firstCdnRecord.Type.String(),
		OriginProtocol: firstCdnRecord.OriginProtocol.String(),
		HTTPPort:       firstCdnRecord.HTTPPort,
		HTTPSPort:      firstCdnRecord.HTTPSPort,
		IsActive:       firstCdnRecord.IsActive,
		LastRefreshAt:  firstCdnRecord.LastRefreshAt,
	}
	if ipGroup, err := firstCdnRecord.Edges.IPGroupOrErr(); err == nil {
		response.IPGroupID = ipGroup.ID
		response.IPGroupName = ipGroup.Name
	}
	successResponse(c, response)
}

// 创建CDN记录
func CreateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateCdnRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建CDN记录
	saveCdnRecord, err := global.GvaDb.CDNRecord.Create().
		SetConfigID(requestParams.ConfigID).
		SetDomain(requestParams.Domain).
		SetType(cdnrecord.Type(requestParams.Type)).
		SetOriginProtocol(cdnrecord.OriginProtocol(requestParams.OriginProtocol)).
		SetHTTPPort(requestParams.HTTPPort).
		SetHTTPSPort(requestParams.HTTPSPort).
		SetIPGroupID(requestParams.IPGroupID).
		Save(c)
	if err != nil {
		fmt.Printf("创建CDN记录: %s\n", err.Error())
		errorResponse(c, http.StatusInternalServerError, "创建CDN记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "创建CDN记录")
	successResponse(c, gin.H{
		"id": saveCdnRecord.ID,
	})
}

// 更新CDN记录
func UpdateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 更新CDN记录
	saveCdnRecord, err := global.GvaDb.CDNRecord.UpdateOne(firstCdnRecord).
		SetOriginProtocol(cdnrecord.OriginProtocol(requestParams.OriginProtocol)).
		SetHTTPPort(requestParams.HTTPPort).
		SetHTTPSPort(requestParams.HTTPSPort).
		SetIPGroupID(requestParams.IPGroupID).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新CDN记录: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新CDN记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, firstCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "更新CDN记录")
	successResponse(c, gin.H{
		"id": saveCdnRecord.ID,
	})
}

// 更新CDN记录状态
func UpdateCdnRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordActiveRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 更新CDN记录状态
	saveCdnRecord, err := global.GvaDb.CDNRecord.UpdateOne(firstCdnRecord).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新CDN记录状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新CDN记录状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, saveCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "更新CDN记录状态")
	successResponse(c, gin.H{
		"id": saveCdnRecord.ID,
	})
}

// 删除CDN记录
func DeleteCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 删除CDN记录
	err = global.GvaDb.CDNRecord.DeleteOne(firstCdnRecord).Exec(c)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除CDN记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "删除CDN记录")
	successResponse(c, gin.H{
		"id": firstCdnRecord.ID,
	})
}
