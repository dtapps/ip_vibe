package gin_server

import (
	"net/http"
	"time"
	"vibe_ip/internal/docs_2"
	"vibe_ip/internal/docs_3"
	"vibe_ip/internal/global"
	"vibe_ip/internal/handlers"
	"vibe_ip/internal/helper/sys_env"

	"github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/wdcbot/qingfeng"
)

// InitRoutes 聚合所有路由注册
func (m *GinServerManager) InitRoutes() {
	m.setupWebRouter()
	m.setupApiRouter()
	m.setupSwaggerRouter()
}

// setupWebRouter 配置 Web 路由
func (m *GinServerManager) setupWebRouter() {
	r := m.engine

	// 根路径 - 提供HTML页面并渲染数据
	r.GET("/", handlers.Dashboard)
	r.GET("/dashboard", handlers.Dashboard)

	// 平台配置页面 - 提供HTML页面并渲染数据
	r.GET("/platform-config", handlers.PlatformConfig)

	// CDN页面 - 提供HTML页面并渲染数据
	r.GET("/cdn", handlers.Cdn)

	// DNS页面 - 提供HTML页面并渲染数据
	r.GET("/dns", handlers.Dns)

	// SSL页面 - 提供HTML页面并渲染数据
	r.GET("/ssl", handlers.Ssl)
	r.GET("/ssl/:id/detail", handlers.SslDetail)
	r.GET("/ssl/:id/history", handlers.SslHistory)

	// SSL导入页面 - 提供HTML页面并渲染数据
	r.GET("/ssl-import", handlers.SslImport)
	r.GET("/ssl-import/:id/detail", handlers.SslImportDetail)

	// TUN页面 - 提供HTML页面并渲染数据
	r.GET("/tun", handlers.Tun)

	// 域名监控页面 - 提供HTML页面并渲染数据
	r.GET("/domain-monitor", handlers.DomainMonitor)
	r.GET("/domain-monitor/:id/detail", handlers.DomainMonitorDetail)
	r.GET("/domain-monitor/:id/history", handlers.DomainMonitorHistory)

	// 设备页面 - 提供HTML页面并渲染数据
	r.GET("/device", handlers.Device)
	r.GET("/device/:id/port-forward/stats", handlers.DevicePortForwardStats)

	// 设备分组页面 - 提供HTML页面并渲染数据
	r.GET("/device-group", handlers.DeviceGroup)
	r.GET("/device-group/:id/port-forward/stats", handlers.DeviceGroupPortForwardStats)

	// IP组页面 - 提供HTML页面并渲染数据
	r.GET("/ip-group", handlers.IPGroup)
	r.GET("/ip-group/:id/port-forward/stats", handlers.IPGroupPortForwardStats)

	// 端口转发页面 - 提供HTML页面并渲染数据
	r.GET("/port-forward", handlers.PortForward)
	r.GET("/port-forward/:id/stats", handlers.PortForwardStats)

	// 通知页面 - 提供HTML页面并渲染数据
	r.GET("/webhook", handlers.Webhook)

	// 通知订阅页面 - 提供HTML页面并渲染数据
	r.GET("/webhook-sub", handlers.WebhookSub)

	// 任务页面 - 提供HTML页面并渲染数据
	r.GET("/task", handlers.Task)

	// 任务页面 - 提供HTML页面并渲染数据
	r.GET("/gocron", handlers.Gocron)

	// 接口页面 - 提供HTML页面并渲染数据
	r.GET("/interface", handlers.Interface)

	// 日志页面 - 提供HTML页面并渲染数据
	r.GET("/log", handlers.Log)

	// 日志内容页面 - 提供HTML页面并渲染数据
	r.GET("/log-content", handlers.LogContent)

	// Cloudflare 隧道管理页面 - 提供HTML页面并渲染数据
	r.GET("/cloudflare-tunnel", handlers.CloudflareTunnel)

	// Cloudflare 站点管理页面 - 提供HTML页面并渲染数据
	r.GET("/cloudflare-zone", handlers.CloudflareZone)
	r.GET("/cloudflare-zone/:id/record", handlers.CloudflareZoneRecord)

	// EdgeOne 站点管理页面 - 提供HTML页面并渲染数据
	r.GET("/edgeone-zone", handlers.EdgeOneZone)
	r.GET("/edgeone-zone/:id/record", handlers.EdgeOneZoneRecord)
	r.GET("/edgeone-zone/:id/origin-group", handlers.EdgeOneZoneOriginGroup)

	// ESA 站点管理页面 - 提供HTML页面并渲染数据
	r.GET("/esa-site", handlers.ESASite)
	r.GET("/esa-site/:id/record", handlers.ESASiteRecord)
	r.GET("/esa-site/:id/origin-pool", handlers.ESASiteOriginPool)

	// 未匹配到任何路由
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, handlers.APIResponse{
			Code:    http.StatusNotFound,
			Message: "path not found",
		})
	})

	// 请求方法不被允许
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, handlers.APIResponse{
			Code:    http.StatusMethodNotAllowed,
			Message: "method not allowed",
		})
	})
}

// setupApiRouter 配置 API 路由
func (m *GinServerManager) setupApiRouter() {

	// 缓存配置
	memoryStore := persist.NewMemoryStore(2 * time.Minute)

	// API路由组
	api := m.engine.Group("/api")
	{
		api.Any("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, handlers.APIResponse{
				Code:    http.StatusOK,
				Message: "api ok",
			})
		})

		// 设置相关API
		settings := api.Group("/settings")
		{
			settings.GET("/", handlers.GetSetting)      // 获取设置
			settings.PATCH("/", handlers.UpdateSetting) // 更新设置
		}

		// 配置Schema相关API
		schemas := api.Group("/schemas")
		{
			schemas.GET("/:config_type/platform-config", cache.CacheByRequestURI(memoryStore, 30*time.Second), handlers.GetPlatformConfigSchema)  // 获取平台配置Schema
			schemas.GET("/:config_type/cdn-record/:config_id", cache.CacheByRequestURI(memoryStore, 30*time.Second), handlers.GetCdnRecordSchema) // 获取CDN记录Schema
			schemas.GET("/:config_type/dns-record/:config_id", cache.CacheByRequestURI(memoryStore, 30*time.Second), handlers.GetDnsRecordSchema) // 获取DNS记录Schema
			schemas.GET("/:config_type/tun-record/:config_id", cache.CacheByRequestURI(memoryStore, 30*time.Second), handlers.GetTunRecordSchema) // 获取TUN记录Schema
		}

		// 平台配置相关API
		platformConfigs := api.Group("/platform-configs")
		{
			platformConfigs.GET("", handlers.GetPlatformConfigs)                 // 获取全部平台配置
			platformConfigs.GET("/:id", handlers.GetPlatformConfigDetail)        // 获取平台配置详情
			platformConfigs.POST("", handlers.CreatePlatformConfig)              // 创建平台配置
			platformConfigs.PATCH("/:id", handlers.UpdatePlatformConfig)         // 更新平台配置
			platformConfigs.DELETE("/:id", handlers.DeletePlatformConfig)        // 删除平台配置
			platformConfigs.DELETE("/batch", handlers.BatchDeletePlatformConfig) // 批量删除平台配置
		}

		// CDN记录管理相关API
		cdnRecords := api.Group("/cdn-records")
		{
			cdnRecords.GET("", handlers.GetCdnRecords)                             // 获取全部CDN记录
			cdnRecords.GET("/:id", handlers.GetCdnRecordDetail)                    // 获取CDN记录详情
			cdnRecords.POST("", handlers.CreateCdnRecord)                          // 创建CDN记录
			cdnRecords.PATCH("/:id", handlers.UpdateCdnRecord)                     // 更新CDN记录
			cdnRecords.PATCH("/:id/active", handlers.UpdateCdnRecordActive)        // 更新CDN记录启用状态
			cdnRecords.PATCH("/batch/active", handlers.BatchUpdateCdnRecordActive) // 批量更新CDN记录启用状态
			cdnRecords.DELETE("/:id", handlers.DeleteCdnRecord)                    // 删除CDN记录
			cdnRecords.DELETE("/batch", handlers.BatchDeleteCdnRecord)             // 批量删除CDN记录
			cdnRecords.GET("/:id/health-check", handlers.HealthCheckCdnRecord)     // 检查CDN记录健康状态
		}

		// DNS记录管理相关API
		dnsRecords := api.Group("/dns-records")
		{
			dnsRecords.GET("", handlers.GetDnsRecords)                             // 获取全部DNS记录
			dnsRecords.GET("/:id", handlers.GetDnsRecordDetail)                    // 获取DNS记录详情
			dnsRecords.POST("", handlers.CreateDnsRecord)                          // 创建DNS记录
			dnsRecords.PATCH("/:id", handlers.UpdateDnsRecord)                     // 更新DNS记录
			dnsRecords.PATCH("/:id/active", handlers.UpdateDnsRecordActive)        // 更新DNS记录启用状态
			dnsRecords.PATCH("/batch/active", handlers.BatchUpdateDnsRecordActive) // 批量更新DNS记录启用状态
			dnsRecords.DELETE("/:id", handlers.DeleteDnsRecord)                    // 删除DNS记录
			dnsRecords.DELETE("/batch", handlers.BatchDeleteDnsRecord)             // 批量删除DNS记录
			dnsRecords.GET("/:id/health-check", handlers.HealthCheckDnsRecord)     // 检查DNS记录健康状态
		}

		// SSL记录管理相关API
		sslRecords := api.Group("/ssl-records")
		{
			sslRecords.GET("", handlers.GetSSLRecords)                                     // 获取全部SSL记录
			sslRecords.GET("/:id", handlers.GetSSLRecordDetail)                            // 获取SSL记录详情
			sslRecords.POST("/import", handlers.ImportSSLRecord)                           // 导入SSL记录
			sslRecords.POST("", handlers.CreateSSLRecord)                                  // 创建SSL记录
			sslRecords.PATCH("/:id", handlers.UpdateSSLRecord)                             // 更新SSL记录
			sslRecords.PATCH("/:id/auto-renewal", handlers.UpdateSSLRecordAutoRenewal)     // 更新SSL记录自动续费状态
			sslRecords.PATCH("/:id/active", handlers.UpdateSSLRecordActive)                // 更新SSL记录启用状态
			sslRecords.PATCH("/batch/active", handlers.BatchUpdateSSLRecordActive)         // 批量更新SSL记录启用状态
			sslRecords.DELETE("/:id", handlers.DeleteSSLRecord)                            // 删除SSL记录
			sslRecords.DELETE("/batch", handlers.BatchDeleteSSLRecord)                     // 批量删除SSL记录
			sslRecords.POST("/:id/apply", handlers.ApplySSLRecord)                         // 申请新的SSL证书
			sslRecords.GET("/:id/history/:history_id", handlers.GetSSLRecordHistoryDetail) // 获取SSL记录历史详情
		}

		// SSL记录管理相关API
		sslImports := api.Group("/ssl-imports")
		{
			sslImports.POST("/import", handlers.ImportSSLImport)       // 导入SSL导入记录
			sslImports.POST("", handlers.CreateSSLImport)              // 创建SSL导入记录
			sslImports.DELETE("/:id", handlers.DeleteSSLImport)        // 删除SSL导入记录
			sslImports.DELETE("/batch", handlers.BatchDeleteSSLImport) // 批量删除SSL导入记录
		}

		// TUN记录管理相关API
		tunRecords := api.Group("/tun-records")
		{
			tunRecords.GET("", handlers.GetTunRecords)                             // 获取全部TUN记录
			tunRecords.GET("/:id", handlers.GetTunRecordDetail)                    // 获取TUN记录详情
			tunRecords.POST("", handlers.CreateTunRecord)                          // 创建TUN记录
			tunRecords.PATCH("/:id", handlers.UpdateTunRecord)                     // 更新TUN记录
			tunRecords.PATCH("/:id/active", handlers.UpdateTunRecordActive)        // 更新TUN记录启用状态
			tunRecords.PATCH("/batch/active", handlers.BatchUpdateTunRecordActive) // 批量更新TUN记录启用状态
			tunRecords.DELETE("/:id", handlers.DeleteTunRecord)                    // 删除TUN记录
			tunRecords.DELETE("/batch", handlers.BatchDeleteTunRecord)             // 批量删除TUN记录
		}

		// 域名监控相关API
		domainMonitors := api.Group("/domain-monitors")
		{
			domainMonitors.GET("", handlers.GetDomainMonitors)                             // 获取所有域名监控
			domainMonitors.GET("/:id", handlers.GetDomainMonitorDetail)                    // 获取域名监控详情
			domainMonitors.POST("", handlers.CreateDomainMonitor)                          // 创建域名监控
			domainMonitors.PATCH("/:id", handlers.UpdateDomainMonitor)                     // 更新域名监控
			domainMonitors.PATCH("/:id/active", handlers.UpdateDomainMonitorActive)        // 更新域名监控启用状态
			domainMonitors.PATCH("/batch/active", handlers.BatchUpdateDomainMonitorActive) // 批量更新域名监控启用状态
			domainMonitors.DELETE("/:id", handlers.DeleteDomainMonitor)                    // 删除域名监控
			domainMonitors.DELETE("/batch", handlers.BatchDeleteDomainMonitor)             // 批量删除域名监控
			domainMonitors.POST("/:id/check", handlers.CheckDomainMonitor)                 // 检查域名监控
		}

		// 设备相关API
		devices := api.Group("/devices")
		{
			devices.GET("", handlers.GetDevices)                                       // 获取全部设备
			devices.GET("/:id", handlers.GetDeviceDetail)                              // 获取设备详情
			devices.POST("", handlers.CreateDevice)                                    // 创建设备
			devices.PATCH("/:id", handlers.UpdateDevice)                               // 更新设备
			devices.DELETE("/:id", handlers.DeleteDevice)                              // 删除设备
			devices.DELETE("/batch", handlers.BatchDeleteDevice)                       // 批量删除设备
			devices.GET("/:id/ip-history", handlers.GetDeviceIpHistory)                // 获取设备IP历史记录
			devices.GET("/:id/health-check", handlers.HealthCheckDevice)               // 检查设备健康状态
			devices.GET("/:id/port-forward/stats", handlers.GetDevicePortForwardStats) // 获取设备端口转发统计数据
			devices.GET("/:id/status", handlers.GetDeviceStatus)                       // 获取设备状态
		}

		// 设备组相关API
		deviceGroups := api.Group("/device-groups")
		{
			deviceGroups.GET("", handlers.GetDeviceGroups)                                             // 获取全部设备组
			deviceGroups.GET("/:id", handlers.GetDeviceGroupDetail)                                    // 获取设备组详情
			deviceGroups.POST("", handlers.CreateDeviceGroup)                                          // 创建设备组
			deviceGroups.PATCH("/:id", handlers.UpdateDeviceGroup)                                     // 更新设备组
			deviceGroups.DELETE("/:id", handlers.DeleteDeviceGroup)                                    // 删除设备组
			deviceGroups.DELETE("/batch", handlers.BatchDeleteDeviceGroup)                             // 批量删除设备组
			deviceGroups.GET("/:id/devices", handlers.GetDeviceGroupDevices)                           // 获取设备组设备
			deviceGroups.GET("/:id/health-check", handlers.HealthCheckDeviceGroup)                     // 检查设备组健康状态
			deviceGroups.GET("/:id/port-forward/stats", handlers.GetDeviceGroupDevicePortForwardStats) // 获取设备组设备端口转发统计数据
		}

		// IP组相关API
		ipGroups := api.Group("/ip-groups")
		{
			ipGroups.GET("", handlers.GetIPGroups)                                             // 获取全部IP组
			ipGroups.GET("/:id", handlers.GetIPGroupDetail)                                    // 获取IP组详情
			ipGroups.POST("", handlers.CreateIPGroup)                                          // 创建IP组
			ipGroups.PATCH("/:id", handlers.UpdateIPGroup)                                     // 更新IP组
			ipGroups.DELETE("/:id", handlers.DeleteIPGroup)                                    // 删除IP组
			ipGroups.DELETE("/batch", handlers.BatchDeleteIPGroup)                             // 批量删除IP组
			ipGroups.GET("/:id/devices", handlers.GetIPGroupDevices)                           // 获取IP组设备
			ipGroups.GET("/:id/health-check", handlers.HealthCheckIPGroup)                     // 检查IP组健康状态
			ipGroups.GET("/:id/port-forward/stats", handlers.GetIPGroupDevicePortForwardStats) // 获取IP组设备端口转发统计数据
		}

		// 端口转发相关API
		portForwards := api.Group("/port-forwards")
		{
			portForwards.GET("", handlers.GetPortForwards)                             // 获取全部端口转发
			portForwards.GET("/:id", handlers.GetPortForwardDetail)                    // 获取端口转发详情
			portForwards.POST("", handlers.CreatePortForward)                          // 创建端口转发
			portForwards.PATCH("/:id", handlers.UpdatePortForward)                     // 更新端口转发
			portForwards.PATCH("/:id/forward", handlers.UpdatePortForwardForward)      // 更新端口转发防火墙状态
			portForwards.PATCH("/:id/active", handlers.UpdatePortForwardActive)        // 更新端口转发启用状态
			portForwards.PATCH("/batch/active", handlers.BatchUpdatePortForwardActive) // 批量更新端口转发启用状态
			portForwards.DELETE("/:id", handlers.DeletePortForward)                    // 删除端口转发
			portForwards.DELETE("/batch", handlers.BatchDeletePortForward)             // 批量删除端口转发
			portForwards.GET("/:id/health-check", handlers.HealthCheckPortForward)     // 检查端口转发健康状态
			portForwards.GET("/:id/stats", handlers.GetPortForwardStats)               // 获取端口转发统计数据
		}

		// 通知相关API
		webhooks := api.Group("/webhooks")
		{
			webhooks.GET("", handlers.GetWebhooks)                             // 获取全部通知
			webhooks.GET("/:id", handlers.GetWebhookDetail)                    // 获取通知详情
			webhooks.POST("", handlers.CreateWebhook)                          // 创建通知
			webhooks.PATCH("/:id", handlers.UpdateWebhook)                     // 更新通知
			webhooks.PATCH("/:id/active", handlers.UpdateWebhookActive)        // 更新通知启用状态
			webhooks.PATCH("/batch/active", handlers.BatchUpdateWebhookActive) // 批量更新通知启用状态
			webhooks.DELETE("/:id", handlers.DeleteWebhook)                    // 删除通知
			webhooks.DELETE("/batch", handlers.BatchDeleteWebhook)             // 批量删除通知
			webhooks.POST("/:id/test", handlers.TestWebhook)                   // 测试通知
		}

		// 通知订阅相关API
		webhookSubs := api.Group("/webhook-subs")
		{
			webhookSubs.GET("", handlers.GetWebhookSubs)                             // 获取全部通知订阅
			webhookSubs.PATCH("/:id/active", handlers.UpdateWebhookSubActive)        // 更新通知订阅启用状态
			webhookSubs.PATCH("/batch/active", handlers.BatchUpdateWebhookSubActive) // 批量更新通知订阅启用状态
			webhookSubs.DELETE("/:id", handlers.DeleteWebhookSub)                    // 删除通知订阅
			webhookSubs.DELETE("/batch", handlers.BatchDeleteWebhookSub)             // 批量删除通知订阅
		}

		// 任务相关API
		tasks := api.Group("/tasks")
		{
			tasks.GET("", handlers.GetTasks)                // 获取全部任务
			tasks.GET("/:id", handlers.GetTaskDetail)       // 获取任务详情
			tasks.POST("/:id", handlers.RunTask)            // 运行任务
			tasks.DELETE("/:id", handlers.DeleteTask)       // 删除任务
			tasks.POST("/clean-all", handlers.CleanAllTask) // 清理所有任务
		}

		// 任务相关API
		gocrons := api.Group("/gocrons")
		{
			gocrons.GET("", handlers.GetGocrons)                            // 获取全部任务
			gocrons.GET("/:id", handlers.GetGocronDetail)                   // 获取任务详情
			gocrons.POST("/:id", handlers.RunGocron)                        // 运行任务
			gocrons.DELETE("/:id", handlers.DeleteGocron)                   // 删除任务
			gocrons.POST("/clean-all", handlers.CleanAllGocron)             // 清理所有任务
			gocrons.POST("/start-scheduler", handlers.StartGocronScheduler) // 启动调度器
			gocrons.POST("/stop-scheduler", handlers.StopGocronScheduler)   // 停止调度器
		}

		// cloudflares 相关API
		cloudflares := api.Group("/cloudflares")
		{
			cloudflares.GET("/:id/domain", cache.CacheByRequestURI(memoryStore, 1*time.Minute), handlers.GetCloudflareDomain) // 获取Cloudflare域名
		}

		// CloudflareTunnel 相关API
		cloudflareTunnels := api.Group("/cloudflare-tunnels")
		{
			cloudflareTunnels.POST("/sync", handlers.SyncCloudflareTunnels) // 同步Cloudflare隧道

			cloudflareTunnels.GET("/:tunnel_id/ingress", handlers.GetCloudflareTunnelIngress) // 获取Cloudflare隧道入站配置

			cloudflareTunnels.GET("/:tunnel_id/ingress/:ingress_id", handlers.GetCloudflareTunnelIngressDetail) // 获取Cloudflare隧道入站配置详情
			cloudflareTunnels.POST("/:tunnel_id/ingress", handlers.CreateCloudflareTunnelIngress)               // 创建Cloudflare隧道入站配置
			cloudflareTunnels.PATCH("/:tunnel_id/ingress/:ingress_id", handlers.UpdateCloudflareTunnelIngress)  // 更新Cloudflare隧道入站配置
			cloudflareTunnels.DELETE("/:tunnel_id/ingress/:ingress_id", handlers.DeleteCloudflareTunnelIngress) // 删除Cloudflare隧道入站配置
		}

		// CloudflareZone 相关API
		cloudflareZones := api.Group("/cloudflare-zones")
		{
			cloudflareZones.POST("/sync", handlers.SyncCloudflareZones) // 同步Cloudflare站点
		}

		// EdgeOneZone 相关API
		edgeoneZones := api.Group("/edgeone-zones")
		{
			edgeoneZones.POST("/sync", handlers.SyncEdgeOneZones) // 同步EdgeOne站点
		}

		// ESASite 相关API
		esaSites := api.Group("/esa-sites")
		{
			esaSites.POST("/sync", handlers.SyncESASites) // 同步ESA站点
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

// setupSwaggerRouter 配置 Swagger 路由
func (m *GinServerManager) setupSwaggerRouter() {

	// Swagger 路由
	// https://github.com/swaggo/swag/blob/v2/README.md
	// https://github.com/swaggo/swag/issues/1588#issuecomment-3360487197
	swaggerGroup := m.engine.Group("/swagger")
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

	// 新 Swagger 路由
	// https://gitee.com/xiaowan1997/qingfeng
	qingfengGroup := m.engine.Group("/qingfeng")
	{
		// 配置 Swagger 页面
		qingfengGroup.GET("/*any", qingfeng.Handler(qingfeng.Config{
			Title:       docs_2.SwaggerInfo.Title,       // 文档标题
			Description: docs_2.SwaggerInfo.Description, // 文档描述
			Version:     docs_2.SwaggerInfo.Version,     // API 版本号
			BasePath:    "/qingfeng",                    // 文档路由前缀
			DocJSON:     docs_2.GetSwaggerBytes(),       // 直接传入 swagger JSON 内容
			EnableDebug: sys_env.GvaEnv.DEBUG,           // 是否启用在线调试
			Environments: []qingfeng.Environment{
				{
					Name:    "本地开发",
					BaseURL: "/api",
				},
			}, // 多环境配置
		}))
	}
}
