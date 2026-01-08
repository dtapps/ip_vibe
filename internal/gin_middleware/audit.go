package gin_middleware

import (
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log/operationlog"

	"github.com/gin-gonic/gin"
)

// AuditContextMiddleware 审计上下文 中间件
func AuditContextMiddleware(source operationlog.Source) gin.HandlerFunc {

	if source == operationlog.SourceUnknown {
		source = operationlog.SourceGin
	}

	// 返回包装后的中间件，添加时间统计
	return func(c *gin.Context) {

		// 设置来源设为 User
		newCtx := schemahook.WithSource(c.Request.Context(), operationlog.SourceUser.String())

		// 将包装后的标准 ctx 回填到 gin.Context 的 Request 中
		// 这样后续所有调用 .Save(ctx) 的地方拿到的都是这个带标识的 ctx
		c.Request = c.Request.WithContext(newCtx)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
