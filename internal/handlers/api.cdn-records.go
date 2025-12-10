package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/backend"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/cdnrecord"
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

// 获取全部CDN记录
//
//	@Description	获取全部CDN记录
//	@Summary		获取全部CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.GetCdnRecordsRequest						false	"请求参数"
//	@Success		200		{object}	APIResponse{data=[]types.GetCdnRecordResponse}	"成功响应"
//	@Failure		400		{object}	APIResponse{}									"参数错误"
//	@Failure		404		{object}	APIResponse{}									"错误响应"
//	@Router			/cdn-records [get]
func GetCdnRecords(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordsRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	records := make([]types.GetCdnRecordResponse, 0)
	// 查询全部CDN记录
	allCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		WhereIf(requestParams.IsActive.Valid, cdnrecord.IsActiveEQ(requestParams.IsActive.Bool)).
		WithConfig().
		WithIPGroup().
		Order(ent.Desc(cdnrecord.FieldCreatedAt)).
		All(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询记录失败: "+err.Error())
		return
	}
	for _, item := range allCdnRecord {
		list := types.GetCdnRecordResponse{
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
		// 查询通知订阅
		list.Webhooks, _ = global.GvaEntDb.WebhookSubscription.Query().
			Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeCdnRecord)).
			Where(webhooksubscription.ObjIDEQ(item.ID)).
			QueryWebhook().
			IDs(c)

		records = append(records, list)
	}
	successResponse(c, records)
}

// 获取CDN记录
//
//	@Description	获取CDN记录
//	@Summary		获取CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64											true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{data=types.GetCdnRecordResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}									请求参数
//	@Failure		404	{object}	APIResponse{}									"错误响应"
//	@Router			/cdn-records/{id} [get]
func GetCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		WithConfig().
		WithIPGroup().
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	response := types.GetCdnRecordResponse{
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
	// 查询通知订阅
	response.Webhooks, _ = global.GvaEntDb.WebhookSubscription.Query().
		Where(webhooksubscription.ObjTypeEQ(webhooksubscription.ObjTypeCdnRecord)).
		Where(webhooksubscription.ObjIDEQ(requestParams.ID)).
		QueryWebhook().
		IDs(c)

	successResponse(c, response)
}

// 创建CDN记录
//
//	@Description	创建CDN记录
//	@Summary		创建CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.CreateCdnRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								请求参数
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records [post]
func CreateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateCdnRecordRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 创建CDN记录
	saveCdnRecord, err := global.GvaEntDb.CDNRecord.Create().
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
	_ = global.GvaOperationLogHelper.RecordDataCreate(c, saveCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "创建CDN记录")

	// 创建通知订阅
	_ = webhook_sub_helper.Create(c, requestParams.WebhookIDs, webhooksubscription.ObjTypeCdnRecord, saveCdnRecord.ID)

	successResponse(c, types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	})
}

// 更新CDN记录
//
//	@Description	更新CDN记录
//	@Summary		更新CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"CDN记录序号"
//	@Param			request	body		types.UpdateCdnRecordRequest				true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id} [patch]
func UpdateCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 更新CDN记录
	saveCdnRecord, err := global.GvaEntDb.CDNRecord.UpdateOne(firstCdnRecord).
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
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, firstCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "更新CDN记录")

	// 更新通知订阅
	_ = webhook_sub_helper.Update(c, requestParams.WebhookIDs, webhooksubscription.ObjTypeCdnRecord, saveCdnRecord.ID)

	successResponse(c, types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	})
}

// 更新CDN记录启用状态
//
//	@Description	更新CDN记录启用状态
//	@Summary		更新CDN记录启用状态
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id		path		int64										true	"CDN记录序号"
//	@Param			request	body		types.UpdateCdnRecordActiveRequest			true	"请求参数"
//	@Success		200		{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400		{object}	APIResponse{}								请求参数
//	@Failure		404		{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id}/active [patch]
func UpdateCdnRecordActive(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCdnRecordActiveRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 更新CDN记录状态
	saveCdnRecord, err := global.GvaEntDb.CDNRecord.UpdateOne(firstCdnRecord).
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
	_ = global.GvaOperationLogHelper.RecordDataUpdate(c, saveCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "更新CDN记录状态")
	successResponse(c, types.GetCdnRecordRequest{
		ID: saveCdnRecord.ID,
	})
}

// 删除CDN记录
//
//	@Description	删除CDN记录
//	@Summary		删除CDN记录
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{data=types.GetCdnRecordRequest}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								请求参数
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/cdn-records/{id} [delete]
func DeleteCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}
	// 删除CDN记录
	err = global.GvaEntDb.CDNRecord.DeleteOne(firstCdnRecord).Exec(c)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "删除CDN记录失败: "+err.Error())
		return
	}
	// 记录操作日志
	_ = global.GvaOperationLogHelper.RecordDataDelete(c, firstCdnRecord.ID, operationlog.ObjTypeCdnRecord, operationlog.StatusSuccess, "删除CDN记录")

	// 删除通知订阅
	_ = webhook_sub_helper.Delete(c, webhooksubscription.ObjTypeCdnRecord, firstCdnRecord.ID)

	successResponse(c, types.GetCdnRecordRequest{
		ID: firstCdnRecord.ID,
	})
}

// 检查CDN记录健康状态
//
//	@Description	检查CDN记录健康状态
//	@Summary		检查CDN记录健康状态
//	@Tags			CDN记录
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64										true	"CDN记录序号"
//	@Success		200	{object}	APIResponse{data=[][]backend.ToPairsStruct}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								请求参数
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Failure		500	{object}	APIResponse{}								"服务错误"
//	@Router			/cdn-records/{id}/healthcheck [get]
func HealthcheckCdnRecord(c *gin.Context) {
	// 请求参数
	var requestParams types.HealthcheckCdnRecordRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数")
		return
	}

	// 查询CDN记录
	firstCdnRecord, err := global.GvaEntDb.CDNRecord.Query().
		Where(cdnrecord.IDEQ(requestParams.ID)).
		First(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "查询CDN记录失败: "+err.Error())
		return
	}

	switch firstCdnRecord.OriginProtocol {
	case cdnrecord.OriginProtocolHTTP:
		firstCdnRecord.HTTPSPort = 0
	case cdnrecord.OriginProtocolHTTPS:
		firstCdnRecord.HTTPPort = 0
	}

	// 检查健康状态
	results := make([][]backend.ToPairsStruct, 0)

	// 并发检查
	eg, egCtx := errgroup.WithContext(c)
	eg.SetLimit(10) // 最多 10 个并发

	if firstCdnRecord.Domain != "" {
		// 检测域名是否可通（含 DNS + HTTP + HTTPS）
		eg.Go(func() error {
			healthCheckResult, _ := backend.CheckDomainHealth(egCtx, firstCdnRecord.Domain, 80, 443, []string{})
			results = append(results, healthCheckResult.ToPairsForDomain())
			return nil
		})
	}

	// if firstCdnRecord.CdnDomain != "" {
	// 	// 检测域名是否可通（含 DNS + HTTP + HTTPS）
	// 	eg.Go(func() error {
	// 		healthCheckResult, _ := backend.CheckDomainHealth(egCtx, firstCdnRecord.CdnDomain, 80, 443, []string{})
	// 		results = append(results, healthCheckResult.ToPairsForDomain())
	// 		return nil
	// 	})
	// }

	// 查询Cdn的IP组
	// firstCdnIpGroup, err := firstCdnRecord.QueryIPGroup().
	// 	WithDevices().
	// 	First(egCtx)
	// if err == nil {
	// 	for _, item := range firstCdnIpGroup.Edges.Devices {
	// 		item := item // 创建副本以避免闭包问题
	// 		if item.IPV4 != "" && firstCdnRecord.Type == cdnrecord.TypeA {
	// 			// 检测 IP 是否可通过 HTTP/HTTPS 访问
	// 			eg.Go(func() error {
	// 				healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV4, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
	// 				results = append(results, healthCheckResult.ToPairsForIP())
	// 				return nil
	// 			})
	// 		}
	// 		if item.IPV6 != "" && firstCdnRecord.Type == cdnrecord.TypeAAAA {
	// 			// 检测 IP 是否可通过 HTTP/HTTPS 访问
	// 			eg.Go(func() error {
	// 				healthCheckResult, _ := backend.CheckIPHealth(egCtx, item.IPV6, firstCdnRecord.HTTPPort, firstCdnRecord.HTTPSPort)
	// 				results = append(results, healthCheckResult.ToPairsForIP())
	// 				return nil
	// 			})
	// 		}
	// 	}
	// }

	if err = eg.Wait(); err != nil {
		errorResponse(c, http.StatusInternalServerError, "CDN记录健康状态检查失败: "+err.Error())
		return
	}

	successResponse(c, results)
}
