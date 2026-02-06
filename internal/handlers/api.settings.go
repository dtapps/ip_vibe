package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// 获取设置
//
//	@Description	获取设置
//	@Summary		获取设置
//	@Tags			Setting
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{data=types.GetSettingResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}								"参数错误"
//	@Failure		404	{object}	APIResponse{}								"错误响应"
//	@Router			/settings [get]
func GetSetting(c *gin.Context) {

	// 获取设置
	// setting := db_cache.GetSetting(c, global.GvaEntDb)
	setting := db_cache.GvaSetting

	// 赋值
	response := types.GetSettingResponse{
		DeviceOfflineInvalidTime: setting.DeviceOfflineInvalidTime,
		IPReportIntervalTime:     setting.IPReportIntervalTime,
		PortForwardReloadTime:    setting.PortForwardReloadTime,
		PortForwardReportTime:    setting.PortForwardReportTime,
		CDNRefreshTaskTime:       setting.CDNRefreshTaskTime,
		DNSRefreshTaskTime:       setting.DNSRefreshTaskTime,
		SSLApplyTaskTime:         setting.SSLApplyTaskTime,
		SSLRefreshTaskTime:       setting.SSLRefreshTaskTime,
		HTTPRequestLogRecordTime: setting.HTTPRequestLogRecordTime,
		HistoryDataRetainTime:    setting.HistoryDataRetainTime,
		StatisticDataRetainTime:  setting.StatisticDataRetainTime,
	}

	successResponse(c, response)
}

// 更新设置
//
//	@Description	更新设置
//	@Summary		更新设置
//	@Tags			Setting
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.UpdateSettingRequest	true	"请求参数"
//	@Success		200		{object}	APIResponse{}				"成功响应"
//	@Failure		400		{object}	APIResponse{}				"参数错误"
//	@Failure		404		{object}	APIResponse{}				"错误响应"
//	@Router			/settings [patch]
func UpdateSetting(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateSettingRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 参数校验：当 DeviceOfflineInvalidTime > 0 时，IPReportIntervalTime 的最大值不能大于它
	if requestParams.DeviceOfflineInvalidTime > 0 {
		minInterval, maxInterval := gocron_helper.ParseTaskTimeInterval(requestParams.IPReportIntervalTime)
		intervalTime := minInterval
		if maxInterval != 0 {
			intervalTime = maxInterval
		}
		if intervalTime != 0 {
			if intervalTime > requestParams.DeviceOfflineInvalidTime {
				errorResponse(c,
					http.StatusBadRequest,
					fmt.Sprintf(
						"设备离线失效时间（%d 分钟）必须大于等于 IP 上报最大间隔（%d 分钟），否则设备会被误判为离线",
						requestParams.DeviceOfflineInvalidTime, intervalTime,
					),
				)
				return
			}
		}
	}

	// 批量更新设置
	err := db_cache.UpdateSetting(c, global.GvaEntDb, &db_cache.Setting{
		DeviceOfflineInvalidTime: requestParams.DeviceOfflineInvalidTime,
		IPReportIntervalTime:     requestParams.IPReportIntervalTime,
		PortForwardReloadTime:    requestParams.PortForwardReloadTime,
		PortForwardReportTime:    requestParams.PortForwardReportTime,
		CDNRefreshTaskTime:       requestParams.CDNRefreshTaskTime,
		DNSRefreshTaskTime:       requestParams.DNSRefreshTaskTime,
		SSLApplyTaskTime:         requestParams.SSLApplyTaskTime,
		SSLRefreshTaskTime:       requestParams.SSLRefreshTaskTime,
		HTTPRequestLogRecordTime: requestParams.HTTPRequestLogRecordTime,
		HistoryDataRetainTime:    requestParams.HistoryDataRetainTime,
		StatisticDataRetainTime:  requestParams.StatisticDataRetainTime,
	})
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}
