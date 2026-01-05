package gin_middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// 配置 BasicAuth 中间件
func BasicAuth(logger *slog.Logger, username string, password string) gin.HandlerFunc {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 如果用户名或密码为空，直接返回空的 HandlerFunc
	if username == "" || password == "" {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// 创建 BasicAuth 中间件
	originalMiddleware := gin.BasicAuth(gin.Accounts{
		username: password,
	})

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {
		start := time.Now()

		// 执行原始的 BasicAuth 中间件
		originalMiddleware(c)

		// 计算消耗时间
		duration := time.Since(start)

		// 记录日志
		logger.DebugContext(c, "[BasicAuth] 中间件处理时间",
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
