package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"vibe_ip/internal/core"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gin_helper"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

func Interface(c *gin.Context) {
	// 请求参数
	var requestParams types.InterfaceRequest
	if err := gin_helper.BindQueryWithDefaults(c, &requestParams); err != nil {
		errorResponse(c, http.StatusBadRequest, "无效的请求参数:"+err.Error())
		return
	}

	// 准备页面数据
	pageData := types.InterfaceResponse{
		CurrentProto: requestParams.Proto,

		HttpAddr: global.GvaEnv.HTTP_ADDR,
		GrpcAddr: global.GvaEnv.GRPC_ADDR,
	}
	if global.GvaEnv.BASIC_AUTH_USERNAME != "" && global.GvaEnv.BASIC_AUTH_PASSWORD != "" {
		if requestParams.Proto == "api" {
			pageData.Token = fmt.Sprintf("%s:%s", global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		} else {
			pageData.Token = core.GetBasicAuthToken(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		}
	}

	// 渲染模板
	renderHTML(
		c,
		filepath.Join("pages", "interface.html"),
		types.BasePageData[types.InterfaceResponse]{
			Debug: global.GvaEnv.DEBUG,
			Title: "接口日志",
			Data:  pageData,
		},
	)
}
