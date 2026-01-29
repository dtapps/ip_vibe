package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// 同步全部Cloudflare隧道
//
//	@Description	同步全部Cloudflare隧道
//	@Summary		同步全部Cloudflare隧道
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/sync-all [post]
func SyncAllCloudflareTunnels(c *gin.Context) {

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 调用 同步全部Cloudflare隧道
	err := service.SyncAllCloudflareTunnels(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	// 调用 同步全部Cloudflare隧道入站规则
	err = service.SyncAllCloudflareTunnelIngresses(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}

// 同步Cloudflare隧道
//
//	@Description	同步Cloudflare隧道
//	@Summary		同步Cloudflare隧道
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64			true	"平台配置序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/sync/{id} [post]
func SyncCloudflareTunnels(c *gin.Context) {
	// 请求参数
	var requestParams types.SyncCloudflareTunnelsRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.SyncCloudflareTunnels(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	err = service.SyncCloudflareTunnelIngresses(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
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
//	@Param			id	path		int64														true	"隧道序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetCloudflareTunnelTunnelIngresse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}												"参数错误"
//	@Failure		404	{object}	APIResponse{}												"错误响应"
//	@Router			/cloudflare-tunnels/{id}/ingresses [get]
func GetCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCloudflareTunnelIngressRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 调用 获取Cloudflare隧道全部入站规则
	ingresses, err := service.GetCloudflareTunnelIngresses(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, ingresses)
}

// 删除Cloudflare隧道入站配置
//
//	@Description	删除Cloudflare隧道入站配置
//	@Summary		删除Cloudflare隧道入站配置
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64														true	"隧道序号"
//	@Param			ingress_id	path		int64														true	"入站配置序号"
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}												"参数错误"
//	@Failure		404	{object}	APIResponse{}												"错误响应"
//	@Router			/cloudflare-tunnels/{id}/ingresses/{ingress_id} [delete]
func DeleteCloudflareTunnelIngress(c *gin.Context) {
	// 请求参数
	var requestParams types.DeleteCloudflareTunnelIngressRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
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
		global.GvaHttpLog,
	)

	// 调用 删除Cloudflare隧道入站规则
	err := service.DeleteCloudflareTunnelIngress(c, requestParams.ID, requestParams.IngressID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}
