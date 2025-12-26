package gin_middleware

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func RequestID(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	// 创建 RequestID 中间件
	originalMiddleware := requestid.New()

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {
		start := time.Now()

		// 执行原始的 RequestID 中间件
		originalMiddleware(c)

		// 计算消耗时间
		duration := time.Since(start)

		// 记录消耗时间日志
		if logger != nil {
			logger.Debug("[RequestID] 中间件处理时间",
				slog.Duration("duration", duration),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
		}

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
