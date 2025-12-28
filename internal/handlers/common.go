package handlers

import (
	"fmt"
	"net/http"
	"vibe_ip/internal/global"

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
