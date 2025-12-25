package gin_middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func BasicAuth(logger *slog.Logger, username string, password string) gin.HandlerFunc {

	// 创建 BasicAuth 中间件
	originalMiddleware := gin.BasicAuth(gin.Accounts{
		username: password,
	})

	// 不启用 BasicAuth 时，直接返回空的 HandlerFunc
	if username == "" || password == "" {
		return func(c *gin.Context) {}
	}

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {
		start := time.Now()

		// 执行原始的 BasicAuth 中间件
		originalMiddleware(c)

		// 计算消耗时间
		duration := time.Since(start)

		// 记录消耗时间日志
		if logger != nil {
			logger.Debug("BasicAuth middleware processing time",
				slog.Duration("duration", duration),
				slog.String("origin", c.GetHeader("Origin")),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status", c.Writer.Status()),
			)
		}
	}
}
