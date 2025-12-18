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
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/docs_2"
	"vibe_ip/internal/docs_3"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/grpc_server"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/mcp_prompt"
	"vibe_ip/internal/mcp_resource"
	"vibe_ip/internal/mcp_tool"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	grpcGatewayRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
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

	initialize.InitEntDb(ctx)      // 初始化 业务数据库
	initialize.InitEntLogDb(ctx)   // 初始化 日志数据库
	initialize.InitGoCron()        // 初始化 GoCron 调度器
	initialize.InitMCP()           // 初始化 MCP
	initialize.InitEntRestyLog()   // 初始化 Resty 日志
	initialize.InitEntHttpLog()    // 初始化 HTTP 日志
	initialize.InitThirdParty(ctx) // 初始化 第三方
	initialize.InitSLog("backend") // 初始化 日志

	initialize.InitOperationLogHelper() // 初始化操作日志助手

	// 打印
	slog.Info("服务启动",
		slog.Bool("debug", global.GvaEnv.DEBUG),
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
	)

	// 强制日志颜色化
	gin.ForceConsoleColor()

	// 创建Gin引擎
	router := gin.Default()

	// 配置调试模式
	if global.GvaEnv.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 配置请求ID中间件
	router.Use(requestid.New())

	// 配置Gzip压缩中间件
	router.Use(gzip.Gzip(gzip.DefaultCompression))

	// 配置CORS中间件
	router.Use(cors.New(cors.Config{
		// 动态允许所有来源
		AllowOriginFunc: func(origin string) bool {
			return true // 允许所有来源
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, // 支持所有 HTTP 方法
		AllowHeaders: []string{
			"Content-Type",         // 支持 Content-Type 头部
			"Authorization",        // 支持 Authorization 头部
			"Cookie",               // 支持 Cookie
			"mcp-protocol-version", // 支持 MCP 协议版本头
		},
		ExposeHeaders: []string{
			"X-Auth-Method",  // 暴露给客户端的头部
			"X-RBAC-Checked", // 暴露给客户端的头部
			"X-Request-ID",   // 暴露给客户端的头部
		},
		AllowCredentials: true, // 允许发送凭据（如 Cookie 和 Authorization）
	}))

	// health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 创建一个受保护的路由组
	var ginAccounts gin.Accounts
	if global.GvaEnv.BASIC_AUTH_USERNAME != "" && global.GvaEnv.BASIC_AUTH_PASSWORD != "" {
		ginAccounts = gin.Accounts{
			global.GvaEnv.BASIC_AUTH_USERNAME: global.GvaEnv.BASIC_AUTH_PASSWORD,
		}
	}

	// Web 资源
	web_resource(ctx, router)

	// 配置BasicAuth中间件
	if len(ginAccounts) > 0 {
		router.Use(gin.BasicAuth(ginAccounts))
	}

	// Web 路由
	web_router(ctx, router)

	// API 路由
	api_router(ctx, router)

	// Swagger 路由
	swagger_router(ctx, router)

	// MCP 启动
	mcp_run(ctx, router)

	// gRPC 启动
	go grpc_run(ctx)

	// gRPC-Gateway 启动
	go grpc_gw_run(ctx, router)

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 启动CDN刷新任务
	ipUpdate := service.NewIpUpdateService(global.GvaEntDb, global.GvaGoCron)
	if err := ipUpdate.StartCdnJob(ctx); err != nil {
		slog.Error("启动CDN刷新任务失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 启动域名监控任务
	domainMonitor := task.NewDomainMonitorTask(global.GvaEntDb, global.GvaGoCron)
	if err := domainMonitor.StartJob(ctx); err != nil {
		slog.Error("启动域名监控任务失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 创建 HTTP 服务器实例（用于优雅关闭）
	httpServerInstance = &http.Server{
		Addr:    global.GvaEnv.HTTP_PORT,
		Handler: router,
	}

	// 启动 Gin 服务器
	slog.Info("Gin 服务器正在启动", slog.String("port", global.GvaEnv.HTTP_PORT))

	// 使用 ListenAndServe 在 goroutine 中运行，以便能够捕获关闭信号
	go func() {
		if err := httpServerInstance.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("Gin 服务器启动失败", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	slog.Info("Gin 服务器启动成功", slog.String("port", global.GvaEnv.HTTP_PORT))

	// 等待中断信号进行优雅关闭
	waitForShutdown()
}

// 优雅关闭函数
func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("收到退出信号，开始优雅关闭...")

	// 执行关闭流程
	gracefulShutdown()
}

// 优雅关闭实现
func gracefulShutdown() {
	shutdownOnce.Do(func() {
		// 1. 首先关闭 HTTP 服务器（停止接收新请求）
		slog.Info("开始关闭 HTTP 服务器...")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if httpServerInstance != nil {
			if err := httpServerInstance.Shutdown(ctx); err != nil {
				slog.Error("HTTP 服务器关闭失败", slog.String("error", err.Error()))
			} else {
				slog.Info("HTTP 服务器已关闭")
			}
		}

		// 2. 关闭 gRPC-Gateway 到上游的连接
		slog.Info("开始关闭 gRPC-Gateway 连接...")
		if grpcConnInstance != nil {
			grpcConnInstance.Close()
			slog.Info("gRPC-Gateway 连接已关闭")
		}

		// 3. 通知 gRPC 服务器开始优雅关闭
		close(shutdownChan)

		// 4. 等待一段时间让 goroutine 完成
		time.Sleep(2 * time.Second)

		// 5. 关闭数据库连接和其他资源（现有的 defer 逻辑）
		slog.Info("开始关闭数据库连接...")

		// 注意：这里不能调用原来的 defer，需要手动调用
		initialize.CloseEntDb()      // 关闭 业务数据库
		initialize.CloseEntLogDb()   // 关闭 日志数据库
		initialize.CloseGoCron()     // 关闭 GoCron 调度器
		initialize.CloseThirdParty() // 关闭 第三方

		slog.Info("所有资源已优雅关闭")
	})
}

// 开发环境
func loadTemplates(templatePath string, funcMap template.FuncMap) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		fmt.Printf("开发环境 布局: %s\n", layout)
		fmt.Printf("开发环境 首页: %s\n", index)
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
			fmt.Printf("开发环境 组件: %s\n", name)
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
			fmt.Printf("开发环境 业务页面: %s\n", name)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
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

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件
		files = append(files, p)             // 当前页面
		if global.GvaEnv.DEBUG {
			fmt.Printf("开发环境 业务页面合并: %s\n", strings.Join(files, " -> "))
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
func loadTemplatesFS(efs embed.FS, funcMap template.FuncMap) multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	templatePath := "templates"

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	if global.GvaEnv.DEBUG {
		fmt.Printf("生产环境 布局: %s\n", layout)
		fmt.Printf("生产环境 首页: %s\n", index)
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
			fmt.Printf("生产环境 组件: %s\n", name)
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
			fmt.Printf("生产环境 业务页面: %s\n", name)
		}

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
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

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件

		files = append(files, p) // 当前页面
		if global.GvaEnv.DEBUG {
			fmt.Printf("生产环境 业务页面合并: %s\n", strings.Join(files, " -> "))
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
func web_resource(ctx context.Context, router *gin.Engine) {
	// 创建带 FuncMap 的 template
	funcMap := template.FuncMap{
		"isDebug": func() bool {
			return global.GvaEnv.DEBUG
		},
		// 格式化时间为字符串，若为空则返回 "—"
		"formatTimeOrNil": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format("2006-01-02 15:04:05")
		},
		// 获取当前设备的 Code
		"selfDeviceCode": func() string {
			code, _ := machineid.ID()
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
				return nil, fmt.Errorf("dict expects even number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
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
		router.HTMLRender = loadTemplates("cmd/backend/templates", funcMap)

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
		router.HTMLRender = loadTemplatesFS(embeddedFiles, funcMap)

		// 静态文件服务 - 使用嵌入的文件系统
		staticFS, _ := fs.Sub(embeddedFiles, "static")
		router.StaticFS("/static", http.FS(staticFS))
	}
}

// Web 路由
func web_router(ctx context.Context, router *gin.Engine) {
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

	// IP组页面 - 提供HTML页面并渲染数据
	router.GET("/ip-group", handlers.IPGroup)

	// 端口转发页面 - 提供HTML页面并渲染数据
	router.GET("/port-forward", handlers.PortForward)

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
func api_router(ctx context.Context, router *gin.Engine) {
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
			devices.GET("", handlers.GetDevices)                         // 获取全部设备
			devices.GET("/:id", handlers.GetDeviceDetail)                // 获取设备详情
			devices.POST("", handlers.CreateDevice)                      // 创建设备
			devices.PATCH("/:id", handlers.UpdateDevice)                 // 更新设备
			devices.DELETE("/:id", handlers.DeleteDevice)                // 删除设备
			devices.GET("/:id/ip-history", handlers.GetDeviceIpHistory)  // 获取设备IP历史记录
			devices.GET("/:id/health-check", handlers.HealthCheckDevice) // 检查设备健康状态
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
		}

		// 通知订阅相关API
		webhookSubs := api.Group("/webhook-subs")
		{
			webhookSubs.GET("", handlers.GetWebhookSubs)                      // 获取全部通知订阅
			webhookSubs.GET("/:id", handlers.GetWebhookSubDetail)             // 获取通知订阅详情
			webhookSubs.POST("", handlers.CreateWebhookSub)                   // 创建通知订阅
			webhookSubs.PATCH("/:id", handlers.UpdateWebhookSub)              // 更新通知订阅
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

		opens := api.Group("/open")
		{
			opens.POST("/device/:device_code/ip-report", handlers.DeviceReportDeviceIp)  // 设备上报IP
			opens.GET("/device/:device_code/port-forwards", handlers.DevicePortForwards) //获取设备端口转发
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
func swagger_router(ctx context.Context, router *gin.Engine) {
	swaggerGroup := router.Group("/swagger")
	{
		swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/docs_2/doc.json"),
		))
	}
	// doc.json 路由
	docsGroup := router.Group("")
	{
		// 配置 Swagger 信息
		docs_2.SwaggerInfo.Version = global.GvaVersion
		docs_2.SwaggerInfo.Host = ""
		docs_3.SwaggerInfo.Version = global.GvaVersion
		docs_3.SwaggerInfo.Host = ""
		docsGroup.GET("/docs_2/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_2.SwaggerInfo.ReadDoc())
		})
		docsGroup.GET("/docs_3/doc.json", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "application/json")
			ctx.String(200, docs_3.SwaggerInfo.ReadDoc())
		})
	}
}

// MCP 启动
func mcp_run(ctx context.Context, router *gin.Engine) {

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
func grpc_run(ctx context.Context) {
	// 判断是否开启 gRPC 功能
	if global.GvaEnv.GRPC_PORT == "" {
		slog.InfoContext(ctx, "gRPC 服务器不启动，因为没有配置端口！")
		return
	}

	// 创建 gRPC 服务器
	grpcServerInstance = grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	// 创建并注册
	openServer := grpc_server.NewOpenServer()
	openServer.RegisterService(grpcServerInstance)

	// 启动 gRPC 服务器
	var err error
	grpcListenerInstance, err = net.Listen("tcp", global.GvaEnv.GRPC_PORT)
	if err != nil {
		slog.ErrorContext(ctx, "gRPC 服务器启动失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.InfoContext(ctx, "gRPC 服务器正在启动", slog.String("port", global.GvaEnv.GRPC_PORT))

	// 使用 Serve 但不阻塞，通过 channel 监听关闭信号
	go func() {
		if err := grpcServerInstance.Serve(grpcListenerInstance); err != nil {
			slog.ErrorContext(ctx, "gRPC 服务器异常退出", slog.String("error", err.Error()))
		}
		slog.InfoContext(ctx, "gRPC 服务器已停止")
	}()

	<-shutdownChan // 等待关闭信号
	grpcServerInstance.GracefulStop()
	if grpcListenerInstance != nil {
		grpcListenerInstance.Close()
	}
	slog.InfoContext(ctx, "gRPC 服务器已优雅关闭")
}

// gRPC-Gateway 启动
func grpc_gw_run(ctx context.Context, router *gin.Engine) {
	// 判断是否开启 gRPC-Gateway 功能
	if global.GvaEnv.GRPC_PORT == "" {
		slog.Info("gRPC-Gateway 不启动，因为没有配置端口！")
		return
	}

	// 连接上游 gRPC - 提升到全局，不再 defer close
	grpcAddr := "localhost" + global.GvaEnv.GRPC_PORT
	var err error
	grpcConnInstance, err = grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		slog.Error("gRPC-Gateway 连接上游失败（忽略路由注册）", slog.String("error", err.Error()))
		return
	}

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

	// 注册你的服务
	if err := grpc_generated.RegisterOpenServiceHandler(ctx, gwMux, grpcConnInstance); err != nil {
		slog.Error("注册 OpenService 网关路由失败", slog.String("error", err.Error()))
		return
	}

	// 关键：把网关挂到 Gin 的 /grpc 前缀
	router.Group("/grpc").Any("/*any", gin.WrapH(http.StripPrefix("/grpc", gwMux)))
	slog.Info("gRPC-Gateway 已挂载到 Gin", slog.String("prefix", "/grpc"), slog.String("upstream", grpcAddr))
}
