package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

//go:embed templates static
var embeddedFiles embed.FS

var (
	version   = "dev"     // 版本号
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

func main() {

	ctx := context.Background()

	initialize.InitEnt(ctx)        // 初始化 ent
	initialize.InitGoCron()        // 初始化 GoCron 调度器
	initialize.InitRestyLog()      // 初始化 Resty 日志
	initialize.InitThirdParty(ctx) // 初始化 第三方
	initialize.InitSLog("backend") // 初始化 日志
	defer func() {
		initialize.CloseEnt()        // 关闭 ent
		initialize.CloseGoCron()     // 关闭 GoCron 调度器
		initialize.CloseThirdParty() // 关闭 第三方
	}()

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
			"Content-Type",
			"Authorization", // 支持 Authorization 头部
			"Cookie",        // 支持 Cookie
		},
		ExposeHeaders: []string{
			"X-Auth-Method",  // 暴露给客户端的头部
			"X-RBAC-Checked", // 暴露给客户端的头部
			"X-Request-ID",   // 暴露给客户端的头部
		},
		AllowCredentials: true, // 允许发送凭据（如 Cookie 和 Authorization）
	}))

	// 配置BasicAuth中间件
	if global.GvaEnv.BASIC_AUTH_USERNAME != "" && global.GvaEnv.BASIC_AUTH_PASSWORD != "" {
		// 账户映射
		router.Use(gin.BasicAuth(gin.Accounts{
			global.GvaEnv.BASIC_AUTH_USERNAME: global.GvaEnv.BASIC_AUTH_PASSWORD,
		}))
	}

	// 创建带 FuncMap 的 template
	funcMap := template.FuncMap{
		"formatTimeOrNil": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format("2006-01-02 15:04:05")
		},
	}

	// 判断是否为开发模式：version == "dev"
	devMode := global.GvaVersion == "dev"
	if devMode {
		// 自定义分隔符
		router.Delims("{{", "}}")

		// 自定义模板功能
		router.SetFuncMap(funcMap)

		// 开发：直接读本地 templates
		router.LoadHTMLGlob("cmd/backend/templates/*.html")

		// 静态文件服务 - 开发：直接读本地 static
		router.StaticFS("/static", http.Dir("cmd/backend/static"))
	} else {
		// 生产：使用 embed
		templ := template.Must(
			template.New("").Funcs(funcMap).ParseFS(embeddedFiles, "templates/*.html"),
		)
		router.SetHTMLTemplate(templ)

		// 静态文件服务 - 使用嵌入的文件系统
		staticFS, _ := fs.Sub(embeddedFiles, "static")
		router.StaticFS("/static", http.FS(staticFS))
	}

	// 根路径 - 提供HTML页面并渲染数据
	router.GET("/", handlers.Index)

	// API路由组
	api := router.Group("/api")
	{
		// 配置相关API
		configs := api.Group("/configs")
		{
			configs.GET("", handlers.GetConfigs)
			configs.GET("/:id", handlers.GetConfig)
			configs.POST("", handlers.CreateConfig)
			configs.PATCH("/:id", handlers.UpdateConfig)
			configs.DELETE("/:id", handlers.DeleteConfig)
		}

		// CDN记录管理相关API
		cdnRecords := api.Group("/cdn-records")
		{
			cdnRecords.GET("", handlers.GetCdnRecords)
			cdnRecords.GET("/:id", handlers.GetCdnRecord)
			cdnRecords.POST("", handlers.CreateCdnRecord)
			cdnRecords.PATCH("/:id", handlers.UpdateCdnRecord)
			cdnRecords.PATCH("/:id/active", handlers.UpdateCdnRecordActive)
			cdnRecords.DELETE("/:id", handlers.DeleteCdnRecord)
		}

		// DNS记录管理相关API
		dnsRecords := api.Group("/dns-records")
		{
			dnsRecords.GET("", handlers.GetDnsRecords)
			dnsRecords.GET("/:id", handlers.GetDnsRecord)
			dnsRecords.POST("", handlers.CreateDnsRecord)
			dnsRecords.PATCH("/:id", handlers.UpdateDnsRecord)
			dnsRecords.PATCH("/:id/active", handlers.UpdateDnsRecordActive)
			dnsRecords.DELETE("/:id", handlers.DeleteDnsRecord)
		}

		// 设备相关API
		devices := api.Group("/devices")
		{
			devices.GET("", handlers.GetDevices)
			devices.GET("/:id", handlers.GetDevice)
			devices.POST("", handlers.CreateDevice)
			devices.PATCH("/:id", handlers.UpdateDevice)
			devices.DELETE("/:id", handlers.DeleteDevice)
			devices.GET("/:id/ip-history", handlers.GetDeviceIpHistory)
			devices.POST("/ip-report", handlers.ReportDeviceIp)               // open api
			devices.GET("/:id/port-forwards", handlers.GetDevicePortForwards) // open api
		}

		// IP组相关API
		ipGroups := api.Group("/ip-groups")
		{
			ipGroups.GET("", handlers.GetIPGroups)
			ipGroups.GET("/:id", handlers.GetIPGroup)
			ipGroups.POST("", handlers.CreateIPGroup)
			ipGroups.PATCH("/:id", handlers.UpdateIPGroup)
			ipGroups.DELETE("/:id", handlers.DeleteIPGroup)
			ipGroups.GET("/:id/devices", handlers.GetIPGroupDevices)
		}

		// 端口转发相关API
		portForwards := api.Group("/port-forwards")
		{
			portForwards.GET("", handlers.GetPortForwards)
			portForwards.GET("/:id", handlers.GetPortForward)
			portForwards.POST("", handlers.CreatePortForward)
			portForwards.PATCH("/:id", handlers.UpdatePortForward)
			portForwards.PATCH("/:id/forward", handlers.UpdatePortForwardForward)
			portForwards.PATCH("/:id/active", handlers.UpdatePortForwardActive)
			portForwards.DELETE("/:id", handlers.DeletePortForward)
		}

		// 通知相关API
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", handlers.GetWebhooks)
			webhooks.GET("/:id", handlers.GetWebhook)
			webhooks.POST("", handlers.CreateWebhook)
			webhooks.PATCH("/:id", handlers.UpdateWebhook)
			webhooks.PATCH("/:id/active", handlers.UpdateWebhookActive)
			webhooks.DELETE("/:id", handlers.DeleteWebhook)
		}

		// 通知订阅相关API
		webhookSubscriptions := api.Group("/webhook-subscriptions")
		{
			webhookSubscriptions.GET("", handlers.GetWebhookSubscriptions)
			webhookSubscriptions.GET("/:id", handlers.GetWebhookSubscription)
			webhookSubscriptions.POST("", handlers.CreateWebhookSubscription)
			webhookSubscriptions.PATCH("/:id", handlers.UpdateWebhookSubscription)
			webhookSubscriptions.PATCH("/:id/active", handlers.UpdateWebhookSubscriptionActive)
			webhookSubscriptions.DELETE("/:id", handlers.DeleteWebhookSubscription)
		}

		// 任务相关API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)
			tasks.GET("/:id", handlers.GetTask)
			tasks.POST("/:id", handlers.RunTask)
			tasks.DELETE("/:id", handlers.DeleteTask)
			tasks.POST("/start", handlers.StartTask)
			tasks.POST("/stop", handlers.StopTask)
			tasks.POST("/clean", handlers.CleanTask)
		}
	}

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "DDNS管理平台服务运行正常",
		})
	})

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 启动IP上报任务
	ipRepost := services.NewIpReportService(global.GvaDb, global.GvaGoCron)
	if err := ipRepost.StartJob(ctx); err != nil {
		fmt.Printf("启动IP上报任务失败: %v\n", err)
		os.Exit(1)
	}

	// 启动IP更新任务
	ipUpdate := services.NewIpUpdateService(global.GvaDb, global.GvaGoCron)
	if err := ipUpdate.StartCdnJob(ctx); err != nil {
		fmt.Printf("启动CDN刷新任务失败: %v\n", err)
		os.Exit(1)
	}

	// 启动服务器
	// fmt.Printf("DDNS管理平台启动在 http://localhost%s\n", global.GvaEnv.HTTP_PORT)
	// fmt.Printf("健康检查地址: http://localhost%s/health\n", global.GvaEnv.HTTP_PORT)
	// fmt.Println("按 Ctrl+C 停止服务")

	// 使用Gin的Run方法启动服务器
	if err := router.Run(global.GvaEnv.HTTP_PORT); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
