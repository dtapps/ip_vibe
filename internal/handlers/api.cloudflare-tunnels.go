package handlers

import (
	"net/http"
	"strings"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/provider/task"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 同步Cloudflare隧道
//
//	@Description	同步Cloudflare隧道
//	@Summary		同步Cloudflare隧道
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			request	body		types.SyncCloudflareTunnelsRequest	true	"请求参数"
//	@Success		200		{object}	types.APIResponse{}					"成功响应"
//	@Failure		400		{object}	types.APIResponse{}					"参数错误"
//	@Failure		404		{object}	types.APIResponse{}					"错误响应"
//	@Router			/cloudflare-tunnels/sync [post]
func SyncCloudflareTunnels(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncCloudflareTunnelsRequest
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建 Cloudflare隧道同步 任务
	cloudflareTunnelSyncTask := task.NewCloudflareTunnelSyncTaskProviderRaw(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	if strings.EqualFold(requestParams.Type, "all") {

		// 添加 Cloudflare隧道同步 一次性任务
		err := cloudflareTunnelSyncTask.AddTask(c, global.GvaEntDb, 0, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-config") && requestParams.ConfigID > 0 {

		// 添加 Cloudflare隧道同步 一次性任务
		err := cloudflareTunnelSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, 0)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	if strings.EqualFold(requestParams.Type, "this-tunnel") && requestParams.ConfigID > 0 && requestParams.TunnelID > 0 {

		// 添加 Cloudflare隧道同步 一次性任务
		err := cloudflareTunnelSyncTask.AddTask(c, global.GvaEntDb, requestParams.ConfigID, requestParams.TunnelID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

	}

	successResponse(c, nil)
}

// 获取Cloudflare隧道入站配置
//
//	@Description	获取Cloudflare隧道入站配置
//	@Summary		获取Cloudflare隧道入站配置
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tunnel_id	path		int64																		true	"隧道序号"
//	@Success		200			{object}	types.APIResponse{data=[]types.GetCloudflareTunnelIngressDetailResponse}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}															"参数错误"
//	@Failure		404			{object}	types.APIResponse{}															"错误响应"
//	@Router			/cloudflare-tunnels/{tunnel_id}/ingresses [get]
func GetCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCloudflareTunnelIngressRequest
	requestParams.TunnelID = cast.ToInt64(c.Param("tunnel_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 获取Cloudflare隧道全部入站规则
	ingresses, err := service.GetCloudflareTunnelIngress(c, requestParams.TunnelID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, ingresses)
}

// 获取Cloudflare隧道入站配置详情
//
//	@Description	获取Cloudflare隧道入站配置详情
//	@Summary		获取Cloudflare隧道入站配置详情
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tunnel_id	path		int64																	true	"隧道序号"
//	@Param			ingress_id	path		int64																	true	"入站规则序号"
//	@Success		200			{object}	types.APIResponse{data=types.GetCloudflareTunnelIngressDetailResponse}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}														"参数错误"
//	@Failure		404			{object}	types.APIResponse{}														"错误响应"
//	@Router			/cloudflare-tunnels/{tunnel_id}/ingresses/{ingress_id} [get]
func GetCloudflareTunnelIngressDetail(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCloudflareTunnelIngressDetailRequest
	requestParams.TunnelID = cast.ToInt64(c.Param("tunnel_id"))
	requestParams.IngressID = cast.ToInt64(c.Param("ingress_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 获取Cloudflare隧道入站规则详情
	ingress, err := service.GetCloudflareTunnelIngressDetail(c, requestParams.TunnelID, requestParams.IngressID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, ingress)
}

// 创建Cloudflare隧道入站配置
//
//	@Description	创建Cloudflare隧道入站配置
//	@Summary		创建Cloudflare隧道入站配置
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tunnel_id	path		int64				true	"隧道序号"
//	@Success		200			{object}	types.APIResponse{}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}	"参数错误"
//	@Failure		404			{object}	types.APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/{tunnel_id}/ingresses [post]
func CreateCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.CreateCloudflareTunnelIngressRequest
	requestParams.TunnelID = cast.ToInt64(c.Param("tunnel_id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 创建Cloudflare隧道入站规则
	err := service.CreateCloudflareTunnelIngress(c, requestParams.TunnelID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}

// 更新Cloudflare隧道入站配置
//
//	@Description	更新Cloudflare隧道入站配置
//	@Summary		更新Cloudflare隧道入站配置
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tunnel_id	path		int64				true	"隧道序号"
//	@Param			ingress_id	path		int64				true	"入站规则序号"
//	@Success		200			{object}	types.APIResponse{}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}	"参数错误"
//	@Failure		404			{object}	types.APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/{tunnel_id}/ingresses/{ingress_id} [patch]
func UpdateCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.UpdateCloudflareTunnelIngressRequest
	requestParams.TunnelID = cast.ToInt64(c.Param("tunnel_id"))
	requestParams.IngressID = cast.ToInt64(c.Param("ingress_id"))
	if err := gin_helper.BindJSONWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 更新Cloudflare隧道入站规则
	err := service.UpdateCloudflareTunnelIngress(c, requestParams.TunnelID, requestParams.IngressID, &requestParams)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}

// 删除Cloudflare隧道入站配置
//
//	@Description	删除Cloudflare隧道入站配置
//	@Summary		删除Cloudflare隧道入站配置
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			tunnel_id	path		int64				true	"隧道序号"
//	@Param			ingress_id	path		int64				true	"入站规则序号"
//	@Success		200			{object}	types.APIResponse{}	"成功响应"
//	@Failure		400			{object}	types.APIResponse{}	"参数错误"
//	@Failure		404			{object}	types.APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/{tunnel_id}/ingresses/{ingress_id} [delete]
func DeleteCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteCloudflareTunnelIngressRequest
	requestParams.TunnelID = cast.ToInt64(c.Param("tunnel_id"))
	requestParams.IngressID = cast.ToInt64(c.Param("ingress_id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
	)

	// 调用 删除Cloudflare隧道入站规则
	err := service.DeleteCloudflareTunnelIngress(c, requestParams.TunnelID, requestParams.IngressID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}
