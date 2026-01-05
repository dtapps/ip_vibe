package gin_middleware

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// 配置 Gzip 压缩 中间件
func Gzip(logger *slog.Logger) gin.HandlerFunc {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 创建 Gzip 压缩 中间件，但排除 SVG 文件
	originalMiddleware := gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".svg"}))

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {
		start := time.Now()

		// 执行原始的 Gzip 中间件
		originalMiddleware(c)

		// 计算消耗时间
		duration := time.Since(start)

		// 记录日志
		logger.DebugContext(c, "[Gzip] 中间件处理时间",
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
		)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
