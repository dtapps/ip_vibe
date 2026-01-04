package handlers

import (
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
		CodeSnippet:  getInterfaceCodeSnippet(requestParams.Proto),

		Username: global.GvaEnv.BASIC_AUTH_USERNAME,
		Password: global.GvaEnv.BASIC_AUTH_PASSWORD,
		Token:    core.GetBasicAuthToken(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD),

		HttpAddr:   global.GvaEnv.HTTP_ADDR,
		GrpcAddr:   global.GvaEnv.GRPC_ADDR,
		ThriftAddr: global.GvaEnv.THRIFT_ADDR,
		DubboAddr:  global.GvaEnv.DUBBO_ADDR,
		TarsAddr:   global.GvaEnv.TARS_ADDR,
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

func getInterfaceCodeSnippet(proto string) string {
	switch proto {
	case "api":
		return "// API 接口代码片段"
	case "mcp":
		return "// MCP 接口代码片段"
	case "grpc":
		return "// gRPC 接口代码片段"
	case "thrift":
		return "// Thrift 接口代码片段"
	case "dubbo":
		return "// Dubbo 接口代码片段"
	case "tars":
		return "// Tars 接口代码片段"
	default:
		return "// 未知协议"
	}
}
