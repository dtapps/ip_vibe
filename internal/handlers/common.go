package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/gin-gonic/gin"
)

// API响应结构
type APIResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// API响应辅助函数
func successResponse(c *gin.Context, data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

func errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
	})
}

// HTTP2 server 推送
// https://gin-gonic.com/zh-cn/docs/examples/http2-server-push/
func pusher(c *gin.Context) {
	if pusher := c.Writer.Pusher(); pusher != nil {
		// 推送 CSS 和 JS 文件
		resources := []string{
			"/static/css/style.css",
			"/static/libs/select2/css/select2.min.css",
			"/static/libs/select2-tailwindcss-v4-theme/select2-tailwindcss-theme.min.css",

			"/static/libs/jquery/jquery.min.js",
			"/static/libs/select2/js/select2.min.js",
			"/static/libs/clipboard/clipboard.min.js",
			"/static/js/uuid.js",
			"/static/js/index.js",
		}
		for _, res := range resources {
			if err := pusher.Push(res, nil); err != nil {
				fmt.Printf("推送失败 %s: %v \n", res, err)
			}
		}
	} else {
		if global.GvaEnv.DEBUG {
			fmt.Printf("推送器为空，跳过推送资源 \n")
		}
	}

}

// 通用渲染HTML函数
func renderHTML[T any](c *gin.Context, templateName string, data types.BasePageData[T]) {
	// 获取基础数据
	baseData := getBaseData[T](c)
	baseData.SetDebug(data.Debug)
	baseData.SetTitle(data.Title)
	baseData.SetData(data.Data)
	// 打印调试信息
	if baseData.Debug {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, baseData)
	}
	// 推送
	pusher(c)
	// 渲染HTML
	c.HTML(http.StatusOK, templateName, baseData)
}

// 指定菜单渲染HTML函数
func renderHTMLWithMenu[T any](c *gin.Context, template string, data types.BasePageData[T], activeMenu string) {
	// 获取基础数据
	baseData := getBaseData[T](c)
	baseData.SetDebug(data.Debug)
	baseData.SetActiveMenu(activeMenu)
	baseData.SetTitle(data.Title)
	baseData.SetData(data.Data)
	// 打印调试信息
	if baseData.Debug {
		fmt.Printf("%s pageData: %+v \n", c.Request.URL.Path, baseData)
	}
	// 推送
	pusher(c)
	// 渲染HTML
	c.HTML(http.StatusOK, template, baseData)
}

// 获取基础数据
func getBaseData[T any](c *gin.Context) types.BasePageData[T] {
	return types.BasePageData[T]{
		ActiveMenu: getActiveMenu(c),
	}
}

// 获取激活菜单
func getActiveMenu(c *gin.Context) string {
	// 先看是否手动设置了
	if menu, exists := c.Get("activeMenu"); exists {
		return menu.(string)
	}
	// 根据路径自动判断
	return getMenuByPath(c.Request.URL.Path)
}

// 根据路径获取菜单
func getMenuByPath(path string) string {
	if path == "/" || path == "/dashboard" {
		return "dashboard"
	}
	// 去掉开头的斜杠，用于匹配路由
	if newPath := strings.TrimPrefix(path, "/"); newPath != "" {
		return newPath
	}
	return ""
}
