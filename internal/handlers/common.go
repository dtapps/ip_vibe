package handlers

import (
	"log/slog"
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
func pusher(c *gin.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	if pusher := c.Writer.Pusher(); pusher != nil {
		// 推送 CSS 和 JS 文件
		resources := []string{
			"/static/css/style.css",
			"/static/libs/select2/css/select2.min.css",
			"/static/libs/select2-tailwindcss-v4-theme/select2-tailwindcss-theme.min.css",

			"/static/libs/jquery/jquery.min.js",
			"/static/libs/select2/js/select2.min.js",
			"/static/libs/clipboard/clipboard.min.js",
			"/static/js/api.js",
			"/static/js/index.js",
		}
		for _, res := range resources {
			if err := pusher.Push(res, nil); err != nil {
				logger.ErrorContext(c, "推送失败",
					slog.String("resource", res),
					slog.Any("err", err),
				)
			}
		}
	} else {
		logger.DebugContext(c, "推送器为空，跳过推送资源")
	}
}

// 通用渲染HTML
func renderHTML[T any](c *gin.Context, templateName string, data types.BasePageData[T]) {
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	// 获取基础数据
	baseData := getBaseData[T](c)
	baseData.SetDebug(data.Debug)
	baseData.SetTitle(data.Title)
	baseData.SetData(data.Data)
	// 打印
	logger.DebugContext(c, "通用渲染HTML",
		slog.String("path", c.Request.URL.Path),
		slog.String("template", templateName),
		slog.Any("data", baseData),
	)
	// 推送
	pusher(c, logger)
	// 渲染HTML
	c.HTML(http.StatusOK, templateName, baseData)
}

// 指定菜单渲染HTML
func renderHTMLWithMenu[T any](c *gin.Context, templateName string, data types.BasePageData[T], activeMenu string) {
	// 日志
	logger := slog.Default()
	if global.GvaGinSlog != nil {
		logger = global.GvaGinSlog.WithLogger()
	}

	// 获取基础数据
	baseData := getBaseData[T](c)
	baseData.SetDebug(data.Debug)
	baseData.SetActiveMenu(activeMenu)
	baseData.SetTitle(data.Title)
	baseData.SetData(data.Data)
	// 打印
	logger.DebugContext(c, "指定菜单渲染HTML",
		slog.String("path", c.Request.URL.Path),
		slog.String("activeMenu", activeMenu),
		slog.String("template", templateName),
		slog.Any("data", baseData),
	)
	// 推送
	pusher(c, logger)
	// 渲染HTML
	c.HTML(http.StatusOK, templateName, baseData)
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
