package gin_middleware

import (
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// RequestIDMiddleware RequestID 中间件
func RequestIDMiddleware() gin.HandlerFunc {

	// 创建 RequestID 中间件
	originalMiddleware := requestid.New()

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {

		// 执行原始的 RequestID 中间件
		originalMiddleware(c)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
