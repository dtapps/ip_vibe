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
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"vibe_ip/internal/backend/service"
	"vibe_ip/internal/backend/task"
	"vibe_ip/internal/docs_2"
	"vibe_ip/internal/docs_3"
	"vibe_ip/internal/global"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/mcp_tool"

	"github.com/denisbrodbeck/machineid"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	initialize.InitRestyLog()      // 初始化 Resty 日志
	initialize.InitThirdParty(ctx) // 初始化 第三方
	initialize.InitSLog("backend") // 初始化 日志
	defer func() {
		initialize.CloseEntDb()      // 关闭 业务数据库
		initialize.CloseEntLogDb()   // 关闭 日志数据库
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

	// health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 创建一个受保护的路由组
	var ginAccounts gin.Accounts
	if global.GvaEnv.BASIC_AUTH_USERNAME != "" && global.GvaEnv.BASIC_AUTH_PASSWORD != "" {
		// 账户映射
		ginAccounts = gin.Accounts{
			global.GvaEnv.BASIC_AUTH_USERNAME: global.GvaEnv.BASIC_AUTH_PASSWORD,
		}
	}

	// Swagger 路由组
	// https://github.com/swaggo/swag/blob/v2/README.md
	// https://github.com/swaggo/swag/issues/1588#issuecomment-3360487197
	swaggerGroup := router.Group("/swagger")
	{
		if len(ginAccounts) > 0 {
			swaggerGroup.Use(gin.BasicAuth(ginAccounts))
		}
		swaggerGroup.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
			ginSwagger.URL("/docs_2/doc.json"), // 注意：这个 URL 是 UI 内部请求的地址
		))
	}
	// doc.json 路由
	docsGroup := router.Group("")
	{
		if len(ginAccounts) > 0 {
			docsGroup.Use(gin.BasicAuth(ginAccounts))
		}
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

	// 配置BasicAuth中间件
	if len(ginAccounts) > 0 {
		router.Use(gin.BasicAuth(ginAccounts))
	}

	// 处理 MCP
	if global.GvaMcpServer != nil {
		// 注册 MCP 实现
		mcp_tool.RegisterPlatformConfigTool(global.GvaMcpServer)
		mcp_tool.RegisterCDNRecordTool(global.GvaMcpServer)
		mcp_tool.RegisterDNSRecordTool(global.GvaMcpServer)
		mcp_tool.RegisterDeviceTool(global.GvaMcpServer)
		mcp_tool.RegisterIPGroupTool(global.GvaMcpServer)
		mcp_tool.RegisterPortForwardTool(global.GvaMcpServer)
		mcp_tool.RegisterDomainMonitorTool(global.GvaMcpServer)
		mcp_tool.RegisterWebhookTool(global.GvaMcpServer)
		mcp_tool.RegisterWebhookSubTool(global.GvaMcpServer)
		// 注册 MCP 路由
		httpTransport := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return global.GvaMcpServer
		}, nil)
		router.Any("/mcp", gin.WrapH(http.StripPrefix("/mcp", httpTransport)))
	}

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

	// 根路径 - 提供HTML页面并渲染数据
	router.GET("/", handlers.Dashboard)
	router.GET("/dashboard", handlers.Dashboard)

	// 平台配置页面 - 提供HTML页面并渲染数据
	router.GET("/config", handlers.Config)

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
			cdnRecords.GET("/:id/health-check", handlers.HealthCheckCdnRecord)
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
			dnsRecords.GET("/:id/health-check", handlers.HealthCheckDnsRecord)
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
			devices.GET("/:id/health-check", handlers.HealthCheckDevice)
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
			ipGroups.GET("/:id/health-check", handlers.HealthCheckIPGroup)
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
			portForwards.GET("/:id/health-check", handlers.HealthCheckPortForward)
		}

		// 域名监控相关API
		domainMonitors := api.Group("/domain-monitors")
		{
			domainMonitors.GET("", handlers.GetDomainMonitors)                      // 获取所有域名监控
			domainMonitors.GET("/:id", handlers.GetDomainMonitor)                   // 获取域名监控详情
			domainMonitors.POST("", handlers.CreateDomainMonitor)                   // 创建域名监控
			domainMonitors.PATCH("/:id", handlers.UpdateDomainMonitor)              // 更新域名监控
			domainMonitors.PATCH("/:id/active", handlers.UpdateDomainMonitorActive) // 更新域名监控状态
			domainMonitors.DELETE("/:id", handlers.DeleteDomainMonitor)             // 删除域名监控
			domainMonitors.POST("/:id/check", handlers.CheckDomainMonitor)          // 检查域名监控
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
		webhookSubs := api.Group("/webhook-subs")
		{
			webhookSubs.GET("", handlers.GetWebhookSubs)
			webhookSubs.GET("/:id", handlers.GetWebhookSub)
			webhookSubs.POST("", handlers.CreateWebhookSub)
			webhookSubs.PATCH("/:id", handlers.UpdateWebhookSub)
			webhookSubs.PATCH("/:id/active", handlers.UpdateWebhookSubActive)
			webhookSubs.DELETE("/:id", handlers.DeleteWebhookSub)
		}

		// 任务相关API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)
			tasks.GET("/:id", handlers.GetTask)
			tasks.POST("/:id", handlers.RunTask)
			tasks.DELETE("/:id", handlers.DeleteTask)
			tasks.POST("/clean", handlers.CleanTask)
			tasks.POST("/start-scheduler", handlers.StartScheduler) // 启动调度器
			tasks.POST("/stop-scheduler", handlers.StopScheduler)   // 停止调度器
		}

		opens := api.Group("/open")
		{
			opens.POST("/device/:device_code/ip-report", handlers.DeviceReportDeviceIp)  // open api
			opens.GET("/device/:device_code/port-forwards", handlers.DevicePortForwards) // open api
		}
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 启动CDN刷新任务
	ipUpdate := service.NewIpUpdateService(global.GvaEntDb, global.GvaGoCron)
	if err := ipUpdate.StartCdnJob(ctx); err != nil {
		fmt.Printf("启动CDN刷新任务失败: %v\n", err)
		os.Exit(1)
	}

	// 启动域名监控任务
	domainMonitor := task.NewDomainMonitorTask(global.GvaEntDb, global.GvaGoCron)
	if err := domainMonitor.StartJob(ctx); err != nil {
		fmt.Printf("启动域名监控任务失败: %v\n", err)
		os.Exit(1)
	}

	// 使用Gin的Run方法启动服务器
	if err := router.Run(global.GvaEnv.HTTP_PORT); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
		os.Exit(1)
	}
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
