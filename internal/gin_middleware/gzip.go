package gin_middleware

import (
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// GzipMiddleware Gzip 压缩 中间件
func GzipMiddleware() gin.HandlerFunc {

	// 创建 Gzip 压缩 中间件，但排除 SVG 文件
	originalMiddleware := gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedExtensions([]string{".svg"}))

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {

		// 执行原始的 Gzip 中间件
		originalMiddleware(c)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
