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

// 获取Cloudflare域名
//
//	@Description	获取Cloudflare域名
//	@Summary		获取Cloudflare域名
//	@Tags			Cloudflare
//	@Accept			application/json
//	@Produce		application/json
//	@Param			id	path		int64														true	"隧道序号"
//	@Success		200	{object}	APIResponse{data=[]types.GetCloudflareDomainResponse}	"成功响应"
//	@Failure		400	{object}	APIResponse{}												"参数错误"
//	@Failure		404	{object}	APIResponse{}												"错误响应"
//	@Router			/cloudflares/{id}/domain [get]
func GetCloudflareDomain(c *gin.Context) {
	// 请求参数
	var requestParams types.GetCloudflareDomainRequest
	requestParams.ID = cast.ToInt64(c.Param("id"))
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 创建服务
	service := service.NewCloudflareService(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGinSlog,
		global.GvaHttpLog,
	)

	// 调用 获取Cloudflare隧道全部入站规则
	ingresses, err := service.GetCloudflareDomains(c, requestParams.ID)
	if err != nil {
		errorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	successResponse(c, ingresses)
}
