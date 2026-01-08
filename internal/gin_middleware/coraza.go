package gin_middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"vibe_ip/internal/handlers"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
	"github.com/gin-gonic/gin"
)

// CorazaMiddleware Coraza WAF 中间件
func CorazaMiddleware(logger *slog.Logger, waf coraza.WAF) gin.HandlerFunc {

	// 如果 WAF 实例为空，直接返回空的 HandlerFunc
	if waf == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()

		// 手动为该请求创建事务
		tx := waf.NewTransaction()
		defer tx.ProcessLogging()
		defer tx.Close()

		// 将 Gin 请求桥接到 Coraza（用请求数据填充 tx）
		if err := populateTransaction(tx, c.Request); err != nil {
			// 记录日志
			logger.ErrorContext(c, "[Coraza] WAF 事务填充失败",
				slog.String("err", err.Error()),
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, handlers.APIResponse{
				Code:    http.StatusInternalServerError,
				Message: "WAF 初始化失败",
			})
			return
		}

		// 处理请求头阶段
		if i := tx.ProcessRequestHeaders(); i != nil {
			// 计算消耗时间
			duration := time.Since(start)

			// 记录日志
			logger.WarnContext(c, "[Coraza] 请求头阶段阻塞",
				slog.Duration("duration", duration),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("ruleID", i.RuleID),
				slog.String("action", i.Action),
				slog.String("data", i.Data),
			)
			c.AbortWithStatusJSON(http.StatusForbidden, handlers.APIResponse{
				Code:    http.StatusForbidden,
				Message: "请求被 WAF 阻塞",
				Data: map[string]any{
					"ruleID": i.RuleID,
					"action": i.Action,
					"data":   i.Data,
				},
			})
			return
		}

		// 如果需要处理请求体（例如 POST/PUT；缓冲并处理）
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				// 记录日志
				logger.ErrorContext(c, "[Coraza] 读取请求体失败",
					slog.String("err", err.Error()),
				)
				c.AbortWithStatusJSON(http.StatusBadRequest, handlers.APIResponse{
					Code:    http.StatusBadRequest,
					Message: "无效请求体",
				})
				return
			}
			c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewReader(body)) // 为下游处理程序恢复

			interruption, _, werr := tx.WriteRequestBody(body)
			if werr != nil {
				// 记录日志
				logger.ErrorContext(c, "[Coraza] WAF 写入请求体失败",
					slog.String("err", werr.Error()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, handlers.APIResponse{
					Code:    http.StatusInternalServerError,
					Message: "WAF 体写入失败",
				})
				return
			}
			if interruption != nil {
				// 计算消耗时间
				duration := time.Since(start)

				// 记录日志
				logger.WarnContext(c, "[Coraza] 写入请求体阶段阻塞",
					slog.Duration("duration", duration),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int("ruleID", interruption.RuleID),
					slog.String("action", interruption.Action),
					slog.String("data", interruption.Data),
				)
				c.AbortWithStatusJSON(http.StatusForbidden, handlers.APIResponse{
					Code:    http.StatusForbidden,
					Message: "请求被 WAF 阻塞",
					Data: map[string]any{
						"ruleID": interruption.RuleID,
						"action": interruption.Action,
						"data":   interruption.Data,
					},
				})
				return
			}

			bodyInterruption, perr := tx.ProcessRequestBody()
			if perr != nil {
				// 记录日志
				logger.ErrorContext(c, "[Coraza] 请求体处理失败",
					slog.String("err", perr.Error()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, handlers.APIResponse{
					Code:    http.StatusInternalServerError,
					Message: "WAF 体处理失败",
				})
				return
			}
			if bodyInterruption != nil {
				// 计算消耗时间
				duration := time.Since(start)

				// 记录日志
				logger.WarnContext(c, "[Coraza] 请求体阶段阻塞",
					slog.Duration("duration", duration),
					slog.String("method", c.Request.Method),
					slog.String("path", c.Request.URL.Path),
					slog.Int("ruleID", bodyInterruption.RuleID),
					slog.String("action", bodyInterruption.Action),
					slog.String("data", bodyInterruption.Data),
				)
				c.AbortWithStatusJSON(http.StatusForbidden, handlers.APIResponse{
					Code:    http.StatusForbidden,
					Message: "请求被 WAF 阻塞",
					Data: map[string]any{
						"ruleID": bodyInterruption.RuleID,
						"action": bodyInterruption.Action,
						"data":   bodyInterruption.Data,
					},
				})
				return
			}
		}

		// 无阻塞：继续 Gin 链
		c.Next()

		// 响应后：处理响应阶段（可选，用于日志/响应检查）
		rw := c.Writer // 或使用您的 responseCapture 如果需要体
		statusStr := strconv.Itoa(rw.Status())
		tx.AddResponseHeader("Status", statusStr)

		proto := c.Request.Proto // 如 "HTTP/1.1"
		if i := tx.ProcessResponseHeaders(rw.Status(), proto); i != nil {
			// 记录日志
			logger.WarnContext(c, "[Coraza] 响应头阶段检测到问题",
				slog.Int("ruleID", i.RuleID),
				slog.String("action", i.Action),
				slog.String("data", i.Data),
			)
			// 注意：响应已发送，无法 abort，但可以日志或进一步处理
		}

		// 如果要处理响应体，请在此添加类似请求体的逻辑（包装 c.Writer，写入 tx.WriteResponseBody 等）

		duration := time.Since(start)
		blocked := c.Writer.Status() >= http.StatusBadRequest
		// 记录日志
		logger.DebugContext(c, "[Coraza] 中间件处理时间",
			slog.Duration("duration", duration),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Bool("blocked", blocked),
		)
	}
}

// populateTransaction 是辅助函数：用请求元数据填充事务（头、URI 等）
func populateTransaction(tx types.Transaction, r *http.Request) error {
	// 解析客户端 IP 和端口
	clientIP, clientPortStr, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		clientIP = r.RemoteAddr
		clientPortStr = "0"
	}
	clientPort, _ := strconv.Atoi(clientPortStr)

	// 服务器信息（示例：使用 Host 作为 server，端口 80/443）
	server := r.Host
	serverPort := 80
	if r.TLS != nil {
		serverPort = 443
	}

	// 处理连接
	tx.ProcessConnection(clientIP, clientPort, server, serverPort)

	// 处理 URI、方法和协议
	uri := r.URL.Path
	if r.URL.RawQuery != "" {
		uri += "?" + r.URL.RawQuery
	}
	tx.ProcessURI(uri, r.Method, r.Proto)

	// 设置服务器名称
	tx.SetServerName(r.Host)

	// 添加请求头
	for name, values := range r.Header {
		for _, value := range values {
			tx.AddRequestHeader(name, value)
		}
	}
	tx.AddRequestHeader("Host", r.Host) // 确保 Host 被添加

	return nil
}
