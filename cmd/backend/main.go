package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/docs_2"
	"vibe_ip/internal/docs_3"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_client"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/grpc_interceptor"
	"vibe_ip/internal/grpc_server"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/helper/gin_slog"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/mcp_prompt"
	"vibe_ip/internal/mcp_resource"
	"vibe_ip/internal/mcp_tool"

	"github.com/denisbrodbeck/machineid"
	"github.com/dromara/carbon/v2"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	grpcGatewayRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wdcbot/qingfeng"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// 全局资源管理
var (
	grpcServerInstance   *grpc.Server
	grpcListenerInstance net.Listener
	httpServerInstance   *http.Server
	grpcConnInstance     *grpc.ClientConn
	shutdownOnce         sync.Once
	shutdownChan         = make(chan struct{})
)

//go:embed templates static
var embeddedFiles embed.FS

var (
	version = "dev" // 版本号
	// version   = "v1.0.0"  // 版本号
	commit    = "unknown" // Git 提交 ID
	buildTime = "unknown" // 构建时间
)

// 初始化版本信息到全局变量
func init() {
	global.GvaVersion = version
	global.GvaGitCommit = commit
	global.GvaBuildTime = buildTime
	global.GvaGoVersion = runtime.Version()
	global.GvaOS = runtime.GOOS
	global.GvaArch = runtime.GOARCH
}

// @title			IPVibe
// @version		dev
// @description	IPVibe
// @host			localhost:9999
// @BasePath		/api/
func main() {

	ctx := context.Background()

	initialize.InitSLog("backend")       // 初始化 日志
	initialize.InitEntSLog("backend")    // 初始化 ent 日志
	initialize.InitGinSLog("backend")    // 初始化 gin 日志
	initialize.InitGrpcSLog("backend")   // 初始化 gRPC 日志
	initialize.InitGocronSLog("backend") // 初始化 gocron 日志
	initialize.InitMcpSLog("backend")    // 初始化 mcp 日志

	initialize.InitCoraza() // 初始化 Coraza 防火墙

	initialize.InitEntDb(ctx, global.GvaEntSlog.WithLogger())     // 初始化 业务数据库
	initialize.InitEntLogDb(ctx, global.GvaEntSlog.WithLogger())  // 初始化 日志数据库
	initialize.InitGoCron(ctx, global.GvaGocronSlog.WithLogger()) // 初始化 GoCron 调度器
	initialize.InitMCP(ctx, global.GvaMcpSlog.WithLogger())       // 初始化 MCP
	initialize.InitEntRestyLog()                                  // 初始化 Resty 日志
	initialize.InitEntHttpLog()                                   // 初始化 HTTP 日志
	initialize.InitThirdParty(ctx)                                // 初始化 第三方

	initialize.InitOperationLogHelper() // 初始化操作日志助手

	// 打印
	slog.Info("服务启动",
		slog.Bool("debug", global.GvaEnv.DEBUG),
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
		slog.String("os", global.GvaOS),
		slog.String("arch", global.GvaArch),
	)

	if global.GvaGinSlog != nil {
		// 禁用控制台颜色
		gin.DisableConsoleColor()
		// 配置gin日志输出到slog
		gin.DefaultWriter = gin_slog.NewSlogWriter(global.GvaGinSlog.WithLogger())
		gin.DefaultErrorWriter = gin_slog.NewSlogWriter(global.GvaGinSlog.WithLogger())
	} else {
		// 强制控制台颜色
		gin.ForceConsoleColor()
	}

	// 创建Gin引擎
	router := gin.Default()

	// 配置调试模式
	if global.GvaEnv.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 配置 请求ID 中间件
	router.Use(gin_middleware.RequestID(
		global.GvaGinSlog.WithLogger(),
	))

	// 配置 Coraza 中间件
	router.Use(gin_middleware.Coraza(
		global.GvaCoraza,
		global.GvaGinSlog.WithLogger(),
	))

	// 配置 Gzip 压缩 中间件
	router.Use(gin_middleware.Gzip(
		global.GvaGinSlog.WithLogger(),
	))

	// 配置 CORS 中间件
	router.Use(gin_middleware.CORS(
		global.GvaGinSlog.WithLogger(),
	))

	// health check
	router.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, handlers.APIResponse{
			Code:    http.StatusOK,
			Message: "ok",
		})
	})

	// Web 资源
	web_resource(ctx, global.GvaGinSlog.WithLogger(), router)

	// 配置 BasicAuth 中间件
	router.Use(gin_middleware.BasicAuth(
		global.GvaGinSlog.WithLogger(),
		global.GvaEnv.BASIC_AUTH_USERNAME,
		global.GvaEnv.BASIC_AUTH_PASSWORD,
	))

	// Web 路由
	web_router(ctx, global.GvaGinSlog.WithLogger(), router)

	// API 路由
	api_router(ctx, global.GvaGinSlog.WithLogger(), router)

	// Swagger 路由
	swagger_router(ctx, global.GvaGinSlog.WithLogger(), router)

	// 新 Swagger 路由
	qingfeng_router(ctx, global.GvaGinSlog.WithLogger(), router)

	// MCP 启动
	mcp_run(ctx, global.GvaMcpSlog.WithLogger(), router)

	// gRPC 启动
	go grpc_run(ctx, global.GvaGrpcSlog.WithLogger())

	// gRPC-Gateway 启动
	go grpc_gw_run(ctx, global.GvaGrpcSlog.WithLogger(), router)

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 启动 CDN、DNS 刷新任务
	ipUpdate := task.NewIPUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := ipUpdate.StartCdnJob(ctx); err != nil {
		slog.Error("启动CDN、DNS刷新任务失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// 启动 SSL 刷新任务
	sslUpdate := task.NewSSLUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := sslUpdate.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动SSL刷新任务失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// 启动域名监控任务
	domainMonitor := task.NewDomainMonitorTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaHttpLog,
		global.GvaEnv.DEBUG,
	)
	if err := domainMonitor.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动域名监控任务失败",
			slog.Any("err", err),
		)
		os.Exit(1)
	}

	// 创建 HTTP 服务器实例（用于优雅关闭）
	httpServerInstance = &http.Server{
		Addr:    global.GvaEnv.HTTP_ADDR,
		Handler: router,
	}

	// 启动 Gin 服务器
	slog.Info("Gin 服务器正在启动", slog.String("port", global.GvaEnv.HTTP_ADDR))

	// 使用 ListenAndServe 在 goroutine 中运行，以便能够捕获关闭信号
	go func() {
		if err := httpServerInstance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Gin 服务器启动失败",
				slog.Any("err", err),
			)
			os.Exit(1)
		}
	}()

	slog.Info("Gin 服务器启动成功", slog.String("port", global.GvaEnv.HTTP_ADDR))

	// 等待中断信号进行优雅关闭
	waitForShutdown(slog.Default())
}

// 优雅关闭函数
func waitForShutdown(logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("收到退出信号，开始优雅关闭...")

	// 执行关闭流程
	gracefulShutdown(logger)
}

// 优雅关闭实现
func gracefulShutdown(logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	shutdownOnce.Do(func() {
		// 1. 首先关闭 HTTP 服务器（停止接收新请求）
		logger.Info("开始关闭 HTTP 服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if httpServerInstance != nil {
			if err := httpServerInstance.Shutdown(ctx); err != nil {
				logger.Error("HTTP 服务器关闭失败",
					slog.Any("err", err),
				)
			} else {
				logger.Info("HTTP 服务器已关闭")
			}
		}

		// 2. 关闭 gRPC-Gateway 到上游的连接
		logger.Info("开始关闭 gRPC-Gateway 连接...")
		if grpcConnInstance != nil {
			grpcConnInstance.Close()
			logger.Info("gRPC-Gateway 连接已关闭")
		}

		// 3. 通知 gRPC 服务器开始优雅关闭
		close(shutdownChan)

		// 4. 等待一段时间让 goroutine 完成
		time.Sleep(1 * time.Second)

		// 5. 关闭数据库连接和其他资源（现有的 defer 逻辑）
		logger.Info("开始关闭数据库连接...")

		// 6. 关闭 第三方服务
		// initialize.CloseThirdParty()

		// 7. 关闭 GoCron 调度器
		initialize.CloseGoCron()

		// 8. 等待一段时间让 goroutine 完成
		time.Sleep(1 * time.Second)

		// 9. 关闭 业务数据库
		initialize.CloseEntDb()

		// 10. 关闭 日志数据库
		initialize.CloseEntLogDb()

		logger.Info("所有资源已优雅关闭")
	})
}

// 开发环境
func loadTemplates(logger *slog.Logger, templatePath string, funcMap template.FuncMap) multitemplate.Renderer {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	r := multitemplate.NewRenderer()

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		logger.Debug("开发环境",
			slog.String("layout", layout),
			slog.String("index", index),
		)
	}

	// 所有业务页面
	pages, err := filepath.Glob(filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := filepath.Glob(filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		if global.GvaEnv.DEBUG {
			logger.Debug("开发环境",
				slog.String("components", name),
			)
		}
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := os.Stat(path)
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		if global.GvaEnv.DEBUG {
			logger.Debug("开发环境",
				slog.String("pages", name),
			)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件
		files = append(files, p)             // 当前页面
		if global.GvaEnv.DEBUG {
			logger.Debug("开发环境",
				slog.String("page", strings.Join(files, " -> ")),
			)
		}
		r.AddFromFilesFuncs(
			name,
			funcMap,
			files...,
		)
	}

	return r
}

// 生产环境
func loadTemplatesFS(logger *slog.Logger, efs embed.FS, funcMap template.FuncMap) multitemplate.Renderer {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	r := multitemplate.NewRenderer()

	templatePath := "templates"

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		logger.Debug("生产环境",
			slog.String("layout", layout),
			slog.String("index", index),
		)
	}

	// 所有业务页面
	pages, err := fs.Glob(efs, filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := fs.Glob(efs, filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		if global.GvaEnv.DEBUG {
			logger.Debug("生产环境",
				slog.String("components", name),
			)
		}
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := embeddedFiles.Open(path)
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		if global.GvaEnv.DEBUG {
			logger.Debug("生产环境",
				slog.String("pages", name),
			)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件

		files = append(files, p) // 当前页面
		if global.GvaEnv.DEBUG {
			logger.Debug("生产环境",
				slog.String("page", strings.Join(files, " -> ")),
			)
		}
		r.AddFromFSFuncs(
			name,
			funcMap,
			efs,
			files...,
		)
	}

	return r
}

// Web 资源
func web_resource(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 创建带 FuncMap 的 template
	funcMap := template.FuncMap{
		// 判断是否为调试模式
		"isDebug": func() bool {
			return global.GvaEnv.DEBUG
		},
		// 格式化时间为字符串，若为空则返回 "—"
		"formatTimeOrNil": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format(time.DateTime)
		},
		// 格式化时间为 "X 前" 格式，如 "2 小时前"
		"diffForHumans": func(t time.Time) string {
			return carbon.NewCarbon(t).DiffForHumans()
		},
		// 获取当前设备 Code
		"selfDeviceCode": func() string {
			code, err := machineid.ID()
			if err != nil {
				logger.Error("获取当前设备 Code 失败",
					slog.Any("err", err),
				)
				return "—"
			}
			return code
		},
		// pagination 组件
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"div":   func(a, b int) int { return a / b },
		"toInt": func(i int64) int { return int(i) },
		// switch / icons / button 组件
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				err := fmt.Errorf("期望有偶数个参数")
				logger.Error("dict 函数参数错误",
					slog.Any("err", err),
				)
				return nil, err
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					err := fmt.Errorf("键必须是字符串")
					logger.Error("dict 函数参数错误",
						slog.Any("err", err),
					)
					return nil, err
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		// 字符串拼接
		"join": func(strs ...string) string {
			return strings.Join(strs, "")
		},
	}

	// 判断是否为开发模式：version == "dev"
	devMode := global.GvaVersion == "dev"
	if devMode {
		// 自定义分隔符
		// router.Delims("{{", "}}")

		// 自定义模板功能
		// router.SetFuncMap(funcMap)

		// 开发：直接读本地 templates
		// router.LoadHTMLGlob("cmd/backend/templates/**/*")

		// 开发环境
		router.HTMLRender = loadTemplates(logger, "cmd/backend/templates", funcMap)

		// 静态文件服务 - 开发：直接读本地 static
		router.StaticFS("/static", http.Dir("cmd/backend/static"))
	} else {
		// 生产：使用 embed
		// templ := template.Must(
		// 	template.New("").
		// 		Funcs(funcMap).
		// 		ParseFS(embeddedFiles, "templates/**/*"),
		// )
		// router.SetHTMLTemplate(templ)

		// 生产环境
		router.HTMLRender = loadTemplatesFS(logger, embeddedFiles, funcMap)

		// 静态文件服务 - 使用嵌入的文件系统
		staticFS, _ := fs.Sub(embeddedFiles, "static")
		router.StaticFS("/static", http.FS(staticFS))
	}
}

// Web 路由
func web_router(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 根路径 - 提供HTML页面并渲染数据
	router.GET("/", handlers.Dashboard)
	router.GET("/dashboard", handlers.Dashboard)

	// 平台配置页面 - 提供HTML页面并渲染数据
	router.GET("/platform-config", handlers.PlatformConfig)

	// CDN页面 - 提供HTML页面并渲染数据
	router.GET("/cdn", handlers.Cdn)

	// DNS页面 - 提供HTML页面并渲染数据
	router.GET("/dns", handlers.Dns)

	// 设备页面 - 提供HTML页面并渲染数据
	router.GET("/device", handlers.Device)
	router.GET("/device/:id/port-forward/stats", handlers.DevicePortForwardStats)

	// IP组页面 - 提供HTML页面并渲染数据
	router.GET("/ip-group", handlers.IPGroup)

	// 端口转发页面 - 提供HTML页面并渲染数据
	router.GET("/port-forward", handlers.PortForward)
	router.GET("/port-forward/:id/stats", handlers.PortForwardStats)

	// SSL页面 - 提供HTML页面并渲染数据
	router.GET("/ssl", handlers.Ssl)
	router.GET("/ssl/:id/detail", handlers.SslDetail)
	router.GET("/ssl/:id/history", handlers.SslHistory)

	// 域名监控页面 - 提供HTML页面并渲染数据
	router.GET("/domain-monitor", handlers.DomainMonitor)
	router.GET("/domain-monitor/:id/detail", handlers.DomainMonitorDetail)
	router.GET("/domain-monitor/:id/history", handlers.DomainMonitorHistory)

	// 通知页面 - 提供HTML页面并渲染数据
	router.GET("/webhook", handlers.Webhook)

	// 通知订阅页面 - 提供HTML页面并渲染数据
	router.GET("/webhook-sub", handlers.WebhookSub)

	// 任务页面 - 提供HTML页面并渲染数据
	router.GET("/task", handlers.Task)

	// 日志页面 - 提供HTML页面并渲染数据
	router.GET("/log", handlers.Log)

	// 未匹配到任何路由
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handlers.APIResponse{
			Code:    http.StatusNotFound,
			Message: "path not found",
		})
	})

	// 请求方法不被允许
	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, handlers.APIResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: "method not allowed",
		})
	})
}

// API 路由
func api_router(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// API路由组
	api := router.Group("/api")
	{
		api.Any("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, handlers.APIResponse{
				Code:    http.StatusOK,
				Message: "api ok",
			})
		})

		// 平台配置相关API
		platformConfigs := api.Group("/platform-configs")
		{
			platformConfigs.GET("", handlers.GetPlatformConfigs)          // 获取全部平台配置
			platformConfigs.GET("/:id", handlers.GetPlatformConfigDetail) // 获取平台配置详情
			platformConfigs.POST("", handlers.CreatePlatformConfig)       // 创建平台配置
			platformConfigs.PATCH("/:id", handlers.UpdatePlatformConfig)  // 更新平台配置
			platformConfigs.DELETE("/:id", handlers.DeletePlatformConfig) // 删除平台配置
		}

		// CDN记录管理相关API
		cdnRecords := api.Group("/cdn-records")
		{
			cdnRecords.GET("", handlers.GetCdnRecords)                         // 获取全部CDN记录
			cdnRecords.GET("/:id", handlers.GetCdnRecordDetail)                // 获取CDN记录详情
			cdnRecords.POST("", handlers.CreateCdnRecord)                      // 创建CDN记录
			cdnRecords.PATCH("/:id", handlers.UpdateCdnRecord)                 // 更新CDN记录
			cdnRecords.PATCH("/:id/active", handlers.UpdateCdnRecordActive)    // 更新CDN记录启用状态
			cdnRecords.DELETE("/:id", handlers.DeleteCdnRecord)                // 删除CDN记录
			cdnRecords.GET("/:id/health-check", handlers.HealthCheckCdnRecord) // 检查CDN记录健康状态
		}

		// DNS记录管理相关API
		dnsRecords := api.Group("/dns-records")
		{
			dnsRecords.GET("", handlers.GetDnsRecords)                         // 获取全部DNS记录
			dnsRecords.GET("/:id", handlers.GetDnsRecordDetail)                // 获取DNS记录详情
			dnsRecords.POST("", handlers.CreateDnsRecord)                      // 创建DNS记录
			dnsRecords.PATCH("/:id", handlers.UpdateDnsRecord)                 // 更新DNS记录
			dnsRecords.PATCH("/:id/active", handlers.UpdateDnsRecordActive)    // 更新DNS记录启用状态
			dnsRecords.DELETE("/:id", handlers.DeleteDnsRecord)                // 删除DNS记录
			dnsRecords.GET("/:id/health-check", handlers.HealthCheckDnsRecord) // 检查DNS记录健康状态
		}

		// 设备相关API
		devices := api.Group("/devices")
		{
			devices.GET("", handlers.GetDevices)                                       // 获取全部设备
			devices.GET("/:id", handlers.GetDeviceDetail)                              // 获取设备详情
			devices.POST("", handlers.CreateDevice)                                    // 创建设备
			devices.PATCH("/:id", handlers.UpdateDevice)                               // 更新设备
			devices.DELETE("/:id", handlers.DeleteDevice)                              // 删除设备
			devices.GET("/:id/ip-history", handlers.GetDeviceIpHistory)                // 获取设备IP历史记录
			devices.GET("/:id/health-check", handlers.HealthCheckDevice)               // 检查设备健康状态
			devices.GET("/:id/port-forward/stats", handlers.GetDevicePortForwardStats) // 获取设备端口转发统计数据
		}

		// IP组相关API
		ipGroups := api.Group("/ip-groups")
		{
			ipGroups.GET("", handlers.GetIPGroups)                         // 获取全部IP组
			ipGroups.GET("/:id", handlers.GetIPGroupDetail)                // 获取IP组详情
			ipGroups.POST("", handlers.CreateIPGroup)                      // 创建IP组
			ipGroups.PATCH("/:id", handlers.UpdateIPGroup)                 // 更新IP组
			ipGroups.DELETE("/:id", handlers.DeleteIPGroup)                // 删除IP组
			ipGroups.GET("/:id/devices", handlers.GetIPGroupDevices)       // 获取IP组设备
			ipGroups.GET("/:id/health-check", handlers.HealthCheckIPGroup) // 检查IP组健康状态
		}

		// 端口转发相关API
		portForwards := api.Group("/port-forwards")
		{
			portForwards.GET("", handlers.GetPortForwards)                         // 获取全部端口转发
			portForwards.GET("/:id", handlers.GetPortForwardDetail)                // 获取端口转发详情
			portForwards.POST("", handlers.CreatePortForward)                      // 创建端口转发
			portForwards.PATCH("/:id", handlers.UpdatePortForward)                 // 更新端口转发
			portForwards.PATCH("/:id/forward", handlers.UpdatePortForwardForward)  // 更新端口转发防火墙状态
			portForwards.PATCH("/:id/active", handlers.UpdatePortForwardActive)    // 更新端口转发启用状态
			portForwards.DELETE("/:id", handlers.DeletePortForward)                // 删除端口转发
			portForwards.GET("/:id/health-check", handlers.HealthCheckPortForward) // 检查端口转发健康状态
			portForwards.GET("/:id/stats", handlers.GetPortForwardStats)           // 获取端口转发统计数据
		}

		// SSL记录管理相关API
		sslRecords := api.Group("/ssl-records")
		{
			sslRecords.GET("", handlers.GetSSLRecords)                                     // 获取全部SSL记录
			sslRecords.GET("/:id", handlers.GetSSLRecordDetail)                            // 获取SSL记录详情
			sslRecords.POST("", handlers.CreateSSLRecord)                                  // 创建SSL记录
			sslRecords.PATCH("/:id", handlers.UpdateSSLRecord)                             // 更新SSL记录
			sslRecords.PATCH("/:id/auto-renewal", handlers.UpdateSSLRecordAutoRenewal)     // 更新SSL记录自动续费状态
			sslRecords.PATCH("/:id/active", handlers.UpdateSSLRecordActive)                // 更新SSL记录启用状态
			sslRecords.DELETE("/:id", handlers.DeleteSSLRecord)                            // 删除SSL记录
			sslRecords.POST("/:id/apply", handlers.ApplySSLRecord)                         // 申请新的SSL证书
			sslRecords.GET("/:id/history/:history_id", handlers.GetSSLRecordHistoryDetail) // 获取SSL记录历史详情
		}

		// 域名监控相关API
		domainMonitors := api.Group("/domain-monitors")
		{
			domainMonitors.GET("", handlers.GetDomainMonitors)                      // 获取所有域名监控
			domainMonitors.GET("/:id", handlers.GetDomainMonitorDetail)             // 获取域名监控详情
			domainMonitors.POST("", handlers.CreateDomainMonitor)                   // 创建域名监控
			domainMonitors.PATCH("/:id", handlers.UpdateDomainMonitor)              // 更新域名监控
			domainMonitors.PATCH("/:id/active", handlers.UpdateDomainMonitorActive) // 更新域名监控启用状态
			domainMonitors.DELETE("/:id", handlers.DeleteDomainMonitor)             // 删除域名监控
			domainMonitors.POST("/:id/check", handlers.CheckDomainMonitor)          // 检查域名监控
		}

		// 通知相关API
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", handlers.GetWebhooks)                      // 获取全部通知
			webhooks.GET("/:id", handlers.GetWebhookDetail)             // 获取通知详情
			webhooks.POST("", handlers.CreateWebhook)                   // 创建通知
			webhooks.PATCH("/:id", handlers.UpdateWebhook)              // 更新通知
			webhooks.PATCH("/:id/active", handlers.UpdateWebhookActive) // 更新通知启用状态
			webhooks.DELETE("/:id", handlers.DeleteWebhook)             // 删除通知
			webhooks.POST("/:id/test", handlers.TestWebhook)            // 测试通知
		}

		// 通知订阅相关API
		webhookSubs := api.Group("/webhook-subs")
		{
			webhookSubs.GET("", handlers.GetWebhookSubs)                      // 获取全部通知订阅
			webhookSubs.PATCH("/:id/active", handlers.UpdateWebhookSubActive) // 更新通知订阅启用状态
			webhookSubs.DELETE("/:id", handlers.DeleteWebhookSub)             // 删除通知订阅
		}

		// 任务相关API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)                        // 获取全部任务
			tasks.GET("/:id", handlers.GetTaskDetail)               // 获取任务详情
			tasks.POST("/:id", handlers.RunTask)                    // 运行任务
			tasks.DELETE("/:id", handlers.DeleteTask)               // 删除任务
			tasks.POST("/clean-all", handlers.CleanAllTask)         // 清理所有任务
			tasks.POST("/start-scheduler", handlers.StartScheduler) // 启动调度器
			tasks.POST("/stop-scheduler", handlers.StopScheduler)   // 停止调度器
		}

		// 开放
		opens := api.Group("/open")
		{
			opens.POST("/device/:device_code/ip-report", handlers.DeviceReportDeviceIp)                              // 设备上报IP
			opens.GET("/device/:device_code/port-forwards", handlers.DevicePortForwards)                             // 获取设备端口转发
			opens.POST("/device/:device_code/port-forwards-stats/:port_forward_id", handlers.CreatePortForwardStats) // 创建端口转发统计
		}

		// 分组内兜底：未匹配路径返回 JSON 404（不会与 /api/d 冲突）
		api.Any("/:any", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, handlers.APIResponse{
				Code:    http.StatusNotFound,
				Message: "api path not found",
			})
		})
	}
}

// Swagger 路由
// https://github.com/swaggo/swag/blob/v2/README.md
// https://github.com/swaggo/swag/issues/1588#issuecomment-3360487197
func swagger_router(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	swaggerGroup := router.Group("/swagger")
	{
		// 配置 Swagger 信息
		docs_2.SwaggerInfo.Version = global.GvaVersion
		docs_2.SwaggerInfo.Host = ""
		docs_3.SwaggerInfo.Version = global.GvaVersion
		docs_3.SwaggerInfo.Host = ""
		swaggerGroup.GET("/docs_2/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_2.SwaggerInfo.ReadDoc())
		})
		swaggerGroup.GET("/docs_3/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_3.SwaggerInfo.ReadDoc())
		})
		// 配置 Swagger 页面
		swaggerGroup.GET("/:any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/swagger/docs_2/doc.json"),
		))
	}
}

// 新 Swagger 路由
// https://gitee.com/xiaowan1997/qingfeng
func qingfeng_router(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	qingfengGroup := router.Group("/qingfeng")
	{
		// 配置 Swagger 页面
		qingfengGroup.GET("/*any", qingfeng.Handler(qingfeng.Config{
			Title:       docs_2.SwaggerInfo.Title, // 文档标题
			Description: docs_2.SwaggerInfo.Description,
			Version:     docs_2.SwaggerInfo.Version, // API 版本号
			BasePath:    "/qingfeng",                // 文档路由前缀
			DocJSON:     docs_2.GetSwaggerBytes(),   // 直接传入 swagger JSON 内容
			EnableDebug: true,                       // 是否启用在线调试
			DarkMode:    true,                       // 是否默认深色模式
			UITheme:     qingfeng.ThemeDefault,      // UI 主题风格: ThemeDefault(默认) / ThemeMinimal(简约) / ThemeModern(现代)
			Environments: []qingfeng.Environment{
				{
					Name:    "本地开发",
					BaseURL: "/api",
				},
			}, // 多环境配置
		}))
	}
}

// MCP 启动
func mcp_run(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 判断是否开启 MCP 功能
	if global.GvaMcpServer == nil {
		return
	}

	// 注册 MCP 实现
	mcp_resource.RegisterTimeResources(global.GvaMcpServer)
	mcp_tool.RegisterTimeTools(global.GvaMcpServer)

	mcp_prompt.RegisterPlatformConfigPrompts(global.GvaMcpServer)
	mcp_resource.RegisterPlatformConfigResources(global.GvaMcpServer)
	mcp_tool.RegisterPlatformConfigTools(global.GvaMcpServer)

	mcp_prompt.RegisterCDNRecordPrompts(global.GvaMcpServer)
	mcp_resource.RegisterCDNRecordResources(global.GvaMcpServer)
	mcp_tool.RegisterCDNRecordTools(global.GvaMcpServer)

	mcp_prompt.RegisterDNSRecordPrompts(global.GvaMcpServer)
	mcp_resource.RegisterDNSRecordResources(global.GvaMcpServer)
	mcp_tool.RegisterDNSRecordTools(global.GvaMcpServer)

	mcp_prompt.RegisterDevicePrompts(global.GvaMcpServer)
	mcp_resource.RegisterDeviceResources(global.GvaMcpServer)
	mcp_tool.RegisterDeviceTools(global.GvaMcpServer)

	mcp_prompt.RegisterIPGroupPrompts(global.GvaMcpServer)
	mcp_resource.RegisterIPGroupResources(global.GvaMcpServer)
	mcp_tool.RegisterIPGroupTools(global.GvaMcpServer)

	mcp_prompt.RegisterPortForwardPrompts(global.GvaMcpServer)
	mcp_resource.RegisterPortForwardResources(global.GvaMcpServer)
	mcp_tool.RegisterPortForwardTools(global.GvaMcpServer)

	mcp_prompt.RegisterDomainMonitorPrompts(global.GvaMcpServer)
	mcp_resource.RegisterDomainMonitorResources(global.GvaMcpServer)
	mcp_tool.RegisterDomainMonitorTools(global.GvaMcpServer)

	mcp_prompt.RegisterWebhookPrompts(global.GvaMcpServer)
	mcp_resource.RegisterWebhookResources(global.GvaMcpServer)
	mcp_tool.RegisterWebhookTools(global.GvaMcpServer)

	mcp_prompt.RegisterWebhookSubPrompts(global.GvaMcpServer)
	mcp_resource.RegisterWebhookSubResources(global.GvaMcpServer)
	mcp_tool.RegisterWebhookSubTools(global.GvaMcpServer)

	// 注册 MCP 路由
	httpTransport := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return global.GvaMcpServer
	}, nil)
	router.Any("/mcp/*path", gin.WrapH(http.StripPrefix("/mcp", httpTransport)))
	router.Any("/mcp", gin.WrapH(http.StripPrefix("/mcp", httpTransport)))
}

// gRPC 启动
func grpc_run(ctx context.Context, logger *slog.Logger) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 判断是否开启 gRPC 功能
	if global.GvaEnv.GRPC_ADDR == "" {
		logger.InfoContext(ctx, "gRPC 服务器不启动，因为没有配置端口")
		return
	}

	// 创建 gRPC 服务器
	opts := make([]grpc.ServerOption, 0)
	usis := make([]grpc.UnaryServerInterceptor, 0)
	// 鉴权拦截器
	usis = append(
		usis,
		grpc_interceptor.BasicAuth(
			global.GvaGrpcSlog.WithLogger(),
			global.GvaEnv.BASIC_AUTH_USERNAME,
			global.GvaEnv.BASIC_AUTH_PASSWORD,
		),
	)
	// 日志拦截器
	usis = append(
		usis,
		grpc_interceptor.UnaryLogger(global.GvaGrpcSlog.WithLogger()),
	)
	opts = append(
		opts,
		grpc.ChainUnaryInterceptor(usis...),
	)
	opts = append(
		opts,
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
	)
	opts = append(
		opts,
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)
	grpcServerInstance = grpc.NewServer(opts...)

	// 注册健康检查服务
	healthServer := health.NewServer()
	healthServer.SetServingStatus("grpc.health.v1.Health", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServerInstance, healthServer)

	// 注册业务服务
	openServer := grpc_server.NewOpenServer()
	openServer.RegisterService(grpcServerInstance)

	// 启动 gRPC 服务器
	var err error
	grpcListenerInstance, err = net.Listen("tcp", global.GvaEnv.GRPC_ADDR)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC 服务器启动失败",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "gRPC 服务器正在启动",
		slog.String("port", global.GvaEnv.GRPC_ADDR),
	)

	// 使用 Serve 但不阻塞，通过 channel 监听关闭信号
	go func() {
		if err := grpcServerInstance.Serve(grpcListenerInstance); err != nil {
			logger.ErrorContext(ctx, "gRPC 服务器异常退出",
				slog.Any("error", err),
			)
		}
		logger.InfoContext(ctx, "gRPC 服务器已停止")
	}()

	<-shutdownChan // 等待关闭信号
	grpcServerInstance.GracefulStop()
	if grpcListenerInstance != nil {
		grpcListenerInstance.Close()
	}
	logger.InfoContext(ctx, "gRPC 服务器已优雅关闭")
}

// gRPC-Gateway 启动
func grpc_gw_run(ctx context.Context, logger *slog.Logger, router *gin.Engine) {
	// 日志
	if logger == nil {
		logger = slog.Default()
	}

	// 判断是否开启 gRPC-Gateway 功能
	if global.GvaEnv.GRPC_ADDR == "" {
		logger.InfoContext(ctx, "gRPC-Gateway 不启动，因为没有配置端口")
		return
	}

	// 创建 gRPC 客户端
	grpc_client := grpc_client.NewGRPCClient()
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		// 鉴权
		grpc.WithPerRPCCredentials(grpc_credential.NewBasicAuth(
			global.GvaGrpcSlog.WithLogger(),
			global.GvaEnv.BASIC_AUTH_USERNAME,
			global.GvaEnv.BASIC_AUTH_PASSWORD,
		)),
	)
	// 连接 gRPC 服务器
	err := grpc_client.Dial(global.GvaEnv.GRPC_ADDR)
	if err != nil {
		logger.ErrorContext(ctx, "gRPC-Gateway 连接上游失败（忽略路由注册）",
			slog.Any("error", err),
		)
		return
	}
	// 获取 gRPC 客户端
	grpcConnInstance = grpc_client.GetClient()

	gwMux := grpcGatewayRuntime.NewServeMux(
		grpcGatewayRuntime.WithErrorHandler(func(
			ctx context.Context, mux *grpcGatewayRuntime.ServeMux, marshaler grpcGatewayRuntime.Marshaler,
			w http.ResponseWriter, r *http.Request, err error,
		) {
			st, ok := status.FromError(err)
			code := http.StatusInternalServerError
			if ok {
				switch st.Code() {
				case codes.Unavailable:
					code = http.StatusServiceUnavailable
				case codes.InvalidArgument:
					code = http.StatusBadRequest
				case codes.NotFound:
					code = http.StatusNotFound
				case codes.Canceled:
					code = http.StatusServiceUnavailable
				}
			}
			w.Header().Set("Content-Type", marshaler.ContentType(nil))
			w.WriteHeader(code)
			_ = marshaler.NewEncoder(w).Encode(handlers.APIResponse{
				Code:    code,
				Message: err.Error(),
			})
		}),
	)

	// 注册服务
	if err := grpc_generated.RegisterOpenServiceHandler(ctx, gwMux, grpcConnInstance); err != nil {
		logger.ErrorContext(ctx, "注册 OpenService 网关路由失败",
			slog.Any("error", err),
		)
		return
	}

	// 关键：把网关挂到 Gin 的 /grpc 前缀
	router.Group("/grpc").Any("/*any", gin.WrapH(http.StripPrefix("/grpc", gwMux)))
	logger.InfoContext(ctx, "gRPC-Gateway 已挂载到 Gin",
		slog.String("prefix", "/grpc"),
		slog.String("upstream", global.GvaEnv.GRPC_ADDR),
	)
}
