package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 获取全部DNS记录
func GetDnsRecords(c *gin.Context) {
	records := make([]types.GetDnsRecordsResponse, 0)
	// 查询全部DNS记录
	allDnsRecord, err := global.GvaDb.DNSRecord.Query().
		WithConfig().
		WithIPGroup().
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	for _, item := range allDnsRecord {
		list := types.GetDnsRecordsResponse{
			ID:            item.ID,
			ConfigID:      item.ConfigID,
			Platform:      item.Edges.Config.Platform.String(),
			RecordID:      item.RecordID,
			Domain:        item.Domain,
			Type:          item.Type.String(),
			IsActive:      item.IsActive,
			LastRefreshAt: item.LastRefreshAt,
		}
		if ipGroup, err := item.Edges.IPGroupOrErr(); err == nil {
			list.IPGroupID = ipGroup.ID
			list.IPGroupName = ipGroup.Name
		}
		if cdnRecord, err := item.Edges.CdnRecordOrErr(); err == nil {
			list.CdnRecordID = cdnRecord.ID
			list.CdnRecordCdnDomain = cdnRecord.CdnDomain
		}
		records = append(records, list)
	}
	successResponse(c, records)
}

// 获取DNS记录
func GetDnsRecord(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	response := types.GetDnsRecordsResponse{
		ID:            firstDnsRecord.ID,
		ConfigID:      firstDnsRecord.ConfigID,
		Platform:      firstDnsRecord.Edges.Config.Platform.String(),
		RecordID:      firstDnsRecord.RecordID,
		Domain:        firstDnsRecord.Domain,
		Type:          firstDnsRecord.Type.String(),
		IsActive:      firstDnsRecord.IsActive,
		LastRefreshAt: firstDnsRecord.LastRefreshAt,
	}
	if ipGroup, err := firstDnsRecord.Edges.IPGroupOrErr(); err == nil {
		response.IPGroupID = ipGroup.ID
		response.IPGroupName = ipGroup.Name
	}
	if cdnRecord, err := firstDnsRecord.Edges.CdnRecordOrErr(); err == nil {
		response.CdnRecordID = cdnRecord.ID
	}
	successResponse(c, response)
}

// 创建DNS记录
func CreateDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDnsRecordRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("CreateDnsRecord params: %+v\n", requestParams)
	}
	// 创建DNS记录
	createDnsRecord := global.GvaDb.DNSRecord.Create().
		SetConfigID(requestParams.ConfigID).
		SetDomain(requestParams.Domain).
		SetType(dnsrecord.Type(requestParams.Type))
	if requestParams.IPGroupID > 0 {
		createDnsRecord.SetIPGroupID(requestParams.IPGroupID)
	}
	if requestParams.CdnRecordID > 0 {
		createDnsRecord.SetCdnRecordID(requestParams.CdnRecordID)
	}
	saveDnsRecord, err := createDnsRecord.Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("创建DNS记录: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "创建DNS记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "创建DNS记录")
	successResponse(c, gin.H{
		"id": saveDnsRecord.ID,
	})
}

// 更新DNS记录状态
func UpdateDnsRecordActive(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	var requestParams types.UpdateDnsRecordActiveRequest
	if err := c.ShouldBindJSON(&requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}
	if global.GvaEnv.DEBUG {
		fmt.Printf("UpdateDnsRecordActive params: %+v\n", requestParams)
	}
	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	// 更新DNS记录
	saveRecord, err := global.GvaDb.DNSRecord.UpdateOne(firstDnsRecord).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新记录: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新DNS记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "更新DNS记录状态")
	successResponse(c, gin.H{
		"id": saveRecord.ID,
	})
}

// 删除DNS记录
func DeleteDnsRecord(c *gin.Context) {
	// 请求参数
	id := c.Param("id")
	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.ID(cast.ToInt64(id))).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	// 删除DNS记录
	err = global.GvaDb.DNSRecord.DeleteOne(firstDnsRecord).Exec(c)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除DNS记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "删除DNS记录")
	successResponse(c, gin.H{
		"id": firstDnsRecord.ID,
	})
}
