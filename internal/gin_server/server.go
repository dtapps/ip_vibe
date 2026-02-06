package gin_server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/helper/gin_slog"
	"vibe_ip/internal/types"

	"github.com/corazawaf/coraza/v3"
	"github.com/gin-gonic/gin"
)

// Config Gin 服务所需的配置参数
type Config struct {
	Addr     string
	Username string
	Password string

	Debug   bool
	Version string
}

// GinServerManager Gin 服务管理器
type GinServerManager struct {
	engine *gin.Engine  // Gin 引擎实例
	server *http.Server // HTTP 服务器实例
	logger *slog.Logger // 日志记录器
	config Config       // Gin 服务配置参数
	coraza *coraza.WAF  // Coraza WAF 实例
}

var (
	instance *GinServerManager
	once     sync.Once
)

// GetManager 获取单例（方便优雅关闭时调用）
func NewManager(logger *slog.Logger, cfg Config) *GinServerManager {
	once.Do(func() {
		if logger == nil {
			logger = slog.Default()
		}
		coraza, err := initCoraza(logger)
		if err != nil {
			instance = &GinServerManager{
				logger: logger,
				config: cfg,
			}
		} else {
			instance = &GinServerManager{
				logger: logger,
				config: cfg,
				coraza: &coraza,
			}
		}
	})
	return instance
}

// initServer 初始化 Gin 服务
func (m *GinServerManager) initServer() {

	if m.logger != nil {
		// 禁用控制台颜色
		gin.DisableConsoleColor()
		// 配置gin日志输出到slog
		gin.DefaultWriter = gin_slog.NewSlogWriter(m.logger)
		gin.DefaultErrorWriter = gin_slog.NewSlogWriter(m.logger)
	} else {
		// 强制控制台颜色
		gin.ForceConsoleColor()
	}

	// 配置调试模式
	if m.config.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin引擎
	// m.engine = gin.Default()
	m.engine = gin.New()

	// 配置 日志 中间件
	m.engine.Use(gin.Logger())

	// 配置 恢复 中间件
	m.engine.Use(gin_middleware.RecoveryMiddleware(m.logger, true))

	// 配置 审计上下文 中间件
	m.engine.Use(gin_middleware.AuditContextMiddleware(operationlog.SourceUser))

	// 配置 请求ID 中间件
	m.engine.Use(gin_middleware.RequestIDMiddleware())

	// 配置 Coraza 中间件
	if m.coraza != nil {
		m.engine.Use(gin_middleware.CorazaMiddleware(m.logger, *m.coraza))
	}

	// 配置 Gzip 压缩 中间件
	m.engine.Use(gin_middleware.CORSMiddleware(m.logger))

	// 配置 CORS 中间件
	m.engine.Use(gin_middleware.GzipMiddleware())

	// 健康检查
	m.engine.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, types.APIResponse{
			Code:    http.StatusOK,
			Message: "ok",
		})
	})

}

// GetServer 获取 Gin 服务，方便其他服务使用
func (m *GinServerManager) GetServer() *gin.Engine {
	if m.engine == nil {
		m.initServer()
	}
	return m.engine
}

// Run 启动 Gin 服务
func (m *GinServerManager) Run(ctx context.Context) {

	// 确保 engine 已初始化
	m.GetServer()

	// 日志
	logger := m.logger.With(
		slog.String("addr", m.config.Addr),
	)

	// 检查地址
	if m.config.Addr == "" {
		logger.ErrorContext(ctx, "Gin 服务启动错误，未配置地址")
		return
	}

	m.server = &http.Server{
		Addr:    m.config.Addr,
		Handler: m.engine,
	}

	logger.InfoContext(ctx, "Gin 服务启动成功")

	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.ErrorContext(ctx, "Gin 服务启动失败",
			slog.Any("err", err),
		)
		panic(err)
	}
}

// Stop 优雅关闭
func (m *GinServerManager) Stop(ctx context.Context) {
	if m.server != nil {
		m.logger.InfoContext(ctx, "Gin 服务正在优雅关闭...")
		if err := m.server.Shutdown(ctx); err != nil {
			m.logger.ErrorContext(ctx, "Gin 服务关闭失败",
				slog.Any("err", err),
			)
		}
	}
	m.logger.InfoContext(ctx, "Gin 服务已安全停止")
}
