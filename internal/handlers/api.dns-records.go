package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/ent"
	"vibe_ip/ent/cdnrecord"
	"vibe_ip/ent/dnsrecord"
	"vibe_ip/ent/operationlog"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

// 获取全部DNS记录
func GetDnsRecords(c *gin.Context) {
	records := make([]types.GetDnsRecordResponse, 0)
	// 查询全部DNS记录
	allDnsRecord, err := global.GvaDb.DNSRecord.Query().
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		Order(ent.Desc(dnsrecord.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	for _, item := range allDnsRecord {
		list := types.GetDnsRecordResponse{
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
	var requestParams types.GetDnsRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.IDEQ(requestParams.ID)).
		WithConfig().
		WithIPGroup().
		WithCdnRecord().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	response := types.GetDnsRecordResponse{
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
		response.CdnRecordCdnDomain = cdnRecord.CdnDomain
	}
	successResponse(c, response)
}

// 创建DNS记录
func CreateDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateDnsRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataCreate(c, saveDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "创建DNS记录")
	successResponse(c, gin.H{
		"id": saveDnsRecord.ID,
	})
}

// 更新DNS记录
func UpdateDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDnsRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	// 更新DNS记录
	createDnsRecord := global.GvaDb.DNSRecord.UpdateOne(firstDnsRecord)
	if requestParams.IPGroupID > 0 {
		createDnsRecord.SetIPGroupID(requestParams.IPGroupID)
	}
	if requestParams.CdnRecordID > 0 {
		createDnsRecord.SetCdnRecordID(requestParams.CdnRecordID)
	}
	saveDnsRecord, err := createDnsRecord.Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新DNS记录: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新DNS记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, saveDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "更新DNS记录")
	successResponse(c, gin.H{
		"id": saveDnsRecord.ID,
	})
}

// 更新DNS记录状态
func UpdateDnsRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateDnsRecordActiveRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}
	// 更新DNS记录状态
	saveRecord, err := global.GvaDb.DNSRecord.UpdateOne(firstDnsRecord).
		SetIsActive(requestParams.IsActive).
		Save(c)
	if err != nil {
		if global.GvaEnv.DEBUG {
			fmt.Printf("更新DNS记录状态: %s\n", err.Error())
		}
		errorResponse(c, http.StatusInternalServerError, "更新DNS记录状态失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataUpdate(c, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "更新DNS记录状态")
	successResponse(c, gin.H{
		"id": saveRecord.ID,
	})
}

// 删除DNS记录
func DeleteDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteDnsRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.IDEQ(requestParams.ID)).
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
	_ = global.GvaOperationLogHelper.SetDb(global.GvaDb).RecordDataDelete(c, firstDnsRecord.ID, operationlog.ObjTypeDNSRecord, operationlog.StatusSuccess, "删除DNS记录")
	successResponse(c, gin.H{
		"id": firstDnsRecord.ID,
	})
}

// 检查DNS记录健康状态
func HealthcheckDnsRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthcheckDnsRecordRequest
	if err := gin_helper.BindUriAndJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询DNS记录
	firstDnsRecord, err := global.GvaDb.DNSRecord.Query().
		Where(dnsrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询DNS记录失败: "+err.Error())
		return
	}

	// 检查健康状态
	results := make([][]backend.ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(c)
	eg.SetLimit(10) // 最多 10 个并发

	if firstDnsRecord.Domain != "" {
		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		eg.Go(func() error {
			healthCheckResult, _ := backend.CheckDomainHealth(egCtx, firstDnsRecord.Domain, 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForDomain())
			return nil
		})
	}

	// 查询Dns的IP组
	firstDnsIpGroup, err := firstDnsRecord.QueryIPGroup().
		WithDevices().
		First(egCtx)
	if err == nil {
		for _, item := range firstDnsIpGroup.Edges.Devices {
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
	}

	// 查询CDN记录
	firstCdnRecord, err := firstDnsRecord.QueryCdnRecord().
		First(egCtx)
	if err == nil {
		if firstCdnRecord.CdnDomain != "" {
			// 检测域名是否可通（含 DNS + HTTP + HTTPS）
			eg.Go(func() error {
				healthCheckResult, _ := backend.CheckDomainHealth(egCtx, firstCdnRecord.CdnDomain, 80, 443, []string{})
				results = append(results, healthCheckResult.ToPairsForDomain())
				return nil
			})
		}
		// 查询CDN的IP组
		firstCdnIpGroup, err := firstCdnRecord.QueryIPGroup().
			WithDevices().
			First(egCtx)
		if err == nil {
			switch firstCdnRecord.OriginProtocol {
			case cdnrecord.OriginProtocolHTTP:
				firstCdnRecord.HTTPSPort = 0
			case cdnrecord.OriginProtocolHTTPS:
				firstCdnRecord.HTTPPort = 0
			}
			for _, item := range firstCdnIpGroup.Edges.Devices {
				item := item // 创建副本以避免闭包问题
				if item.IPV4 != "" {
					// 检测 IP 是否可通过 HTTP/HTTPS 访问
					eg.Go(func() error {
						healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV4, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
						results = append(results, healthCheckResult.ToPairsForIP())
						return nil
					})
				}
				if item.IPV6 != "" {
					// 检测 IP 是否可通过 HTTP/HTTPS 访问
					eg.Go(func() error {
						healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV6, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
						results = append(results, healthCheckResult.ToPairsForIP())
						return nil
					})
				}
			}
		}
	}

	if err = eg.Wait(); err != nil {
		errorResponse(c, http.StatusInternalServerError, "DNS记录健康状态检查失败: "+err.Error())
		return
	}

	successResponse(c, results)
}
