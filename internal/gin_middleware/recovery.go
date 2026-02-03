package gin_middleware

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware Recovery 中间件
func RecoveryMiddleware(logger *slog.Logger, stack bool) gin.HandlerFunc {

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查连接是否断开
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取请求详情
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					logger.ErrorContext(c, "[Recovery] Broken pipe",
						slog.Any("err", err),
						slog.String("request", string(httpRequest)),
					)
					c.Error(err.(error))
					c.Abort()
					return
				}

				// 构建日志字段
				fields := []slog.Attr{
					slog.Any("err", err),
					slog.String("request", string(httpRequest)),
				}

				// 如果开启堆栈打印
				if stack {
					fields = append(fields, slog.String("stack", string(debug.Stack())))
				}

				logger.LogAttrs(c, slog.LevelError, "[Recovery] from panic", fields...)

				// 响应 500
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
