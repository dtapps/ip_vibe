package gin_middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/corazawaf/coraza/v3"
	corazahttp "github.com/corazawaf/coraza/v3/http"
	"github.com/gin-gonic/gin"
)

func Coraza(waf coraza.WAF, logger *slog.Logger) gin.HandlerFunc {

	// 这里完全等价于 gin-coraza：用 WrapHandler 包一个 http.Handler
	handler := corazahttp.WrapHandler(waf, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 关键点：什么都不做，只是占位
	}))

	return func(c *gin.Context) {
		start := time.Now()

		// 调用被 Wrap 的 handler
		handler.ServeHTTP(c.Writer, c.Request)

		duration := time.Since(start)

		// 如果 Coraza 触发了 deny，这里通常已经写过状态码
		blocked := c.Writer.Status() >= 400

		if logger != nil {
			logger.Debug("Coraza middleware processing time",
				slog.Duration("duration", duration),
				slog.Bool("blocked", blocked),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status", c.Writer.Status()),
			)
		}

		// 被拦截就中断 Gin
		if blocked {
			c.Abort()
			return
		}

		// 否则继续 Gin 链
		c.Next()
	}
}
