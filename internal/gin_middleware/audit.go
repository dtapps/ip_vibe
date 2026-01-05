package gin_middleware

import (
	"log/slog"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log/operationlog"

	"github.com/gin-gonic/gin"
)

// AuditContextMiddleware 配置 审计上下文中间件 中间件
func AuditContextMiddleware(logger *slog.Logger) gin.HandlerFunc {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 审计上下文中间件
	return func(c *gin.Context) {

		// 设置来源设为 User (因为是 API 请求)
		ctx := schemahook.WithSource(c.Request.Context(), operationlog.SourceUser.String())

		// 将包装后的标准 ctx 回填到 gin.Context 的 Request 中
		// 这样后续所有调用 .Save(ctx) 的地方拿到的都是这个带标识的 ctx
		c.Request = c.Request.WithContext(ctx)

		// 无阻塞：继续 Gin 链
		c.Next()
	}
}
