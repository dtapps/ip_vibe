package gin_middleware

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS(logger *slog.Logger) gin.HandlerFunc {

	// 配置 CORS 中间件
	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// 调试日志
			if logger != nil {
				logger.Info("[CORS] Request from origin",
					slog.String("origin", origin),
				)
			}
			return true // 允许所有来源
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, // 支持所有 HTTP 方法
		AllowHeaders: []string{
			"Content-Type",         // 支持 Content-Type 头部
			"Authorization",        // 支持 Authorization 头部
			"Cookie",               // 支持 Cookie
			"mcp-protocol-version", // 支持 MCP 协议版本头
		},
		ExposeHeaders: []string{
			"X-Auth-Method",  // 暴露给客户端的头部
			"X-RBAC-Checked", // 暴露给客户端的头部
			"X-Request-ID",   // 暴露给客户端的头部
		},
		AllowCredentials: true,  // 允许发送凭据（如 Cookie 和 Authorization）
		MaxAge:           86400, // 可选：设置预检请求的缓存时间
	}

	// 创建 CORS 中间件
	originalMiddleware := cors.New(corsConfig)

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {
		start := time.Now()

		// 执行原始的 CORS 中间件
		originalMiddleware(c)

		// 计算消耗时间
		duration := time.Since(start)

		// 记录消耗时间日志
		if logger != nil {
			logger.Debug("[CORS] 中间件处理时间",
				slog.Duration("duration", duration),
				slog.String("origin", c.GetHeader("Origin")),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
		}

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
