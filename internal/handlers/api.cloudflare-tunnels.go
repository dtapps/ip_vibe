package handlers

import (
	"net/http"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/global"

	"github.com/gin-gonic/gin"
)

// 同步Cloudflare隧道
//
//	@Description	同步Cloudflare隧道
//	@Summary		同步Cloudflare隧道
//	@Tags			CloudflareTunnel
//	@Accept			application/json
//	@Produce		application/json
//	@Success		200	{object}	APIResponse{}	"成功响应"
//	@Failure		400	{object}	APIResponse{}	"参数错误"
//	@Failure		404	{object}	APIResponse{}	"错误响应"
//	@Router			/cloudflare-tunnels/sync [post]
func SyncCloudflareTunnels(c *gin.Context) {

	// 创建服务
	service := service.NewCloudflareTunnelService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 调用
	err := service.SyncCloudflareTunnels(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}
	err = service.SyncCloudflareTunnelIngresses(c)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, nil)
}
