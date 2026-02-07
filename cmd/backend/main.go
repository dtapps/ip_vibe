package main

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
	"vibe_ip/internal/backend/tasks"
	"vibe_ip/internal/ent/schema/schemahook"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/gin_middleware"
	"vibe_ip/internal/gin_server"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_server"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/manager/cron_manager"
	"vibe_ip/internal/manager/mixed_manager"
	"vibe_ip/internal/mcp_server"
)

//go:embed templates static
var embeddedFiles embed.FS

// @title			IPVibe
// @version		dev
// @description	IPVibe
// @host			localhost:9999
// @BasePath		/api/
func main() {

	// 全局 Panic 捕获
	defer func() {
		if r := recover(); r != nil {
			// 获取堆栈信息
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackInfo := string(buf[:n])

			// 格式化错误消息
			errMsg := fmt.Sprintf("%v", r)

			// 使用标准 slog 记录到控制台/文件（此时 global 变量可能不可靠，用标准 slog 最稳）
			slog.Error("🔥 系统触发致命崩溃 (Panic)",
				slog.String("err", errMsg),
				slog.String("stack", stackInfo),
			)

			// 给日志留一点写入时间
			time.Sleep(time.Second)
			os.Exit(1)
		}
	}()

	// 创建上下文
	ctx := schemahook.WithSource(context.Background(), operationlog.SourceSystem.String())

	initialize.InitSLog("backend")       // 初始化 日志
	initialize.InitEntSLog("backend")    // 初始化 ent 日志
	initialize.InitGinSLog("backend")    // 初始化 gin 日志
	initialize.InitGrpcSLog("backend")   // 初始化 gRPC 日志
	initialize.InitGocronSLog("backend") // 初始化 gocron 日志
	initialize.InitMcpSLog("backend")    // 初始化 mcp 日志
	initialize.InitHttpSLog("backend")   // 初始化 http 日志
	initialize.InitRestySLog("backend")  // 初始化 resty 日志
	initialize.InitReqSLog("backend")    // 初始化 req 日志
	initialize.EnsureDefaultSLog()       // 确保日志不为 nil

	initialize.InitEntDb(ctx, global.GvaEntSlog)       // 初始化 业务数据库
	initialize.InitEntLogDb(ctx, global.GvaEntLogSlog) // 初始化 日志数据库
	db_cache.InitSetting(ctx, global.GvaEntDb)         // 初始化 设置
	initialize.InitGoCron(ctx, global.GvaGocronSlog)   // 初始化 GoCron 调度器

	initialize.InitEntHttpLog(global.GvaHttpSlog)   // 初始化 HTTP 日志
	initialize.InitEntRestyLog(global.GvaRestySlog) // 初始化 Resty 日志
	initialize.InitEntReqLog(global.GvaReqSlog)     // 初始化 Req 日志

	initialize.InitThirdParty(ctx, global.GvaSlog, global.GvaHttpLog) // 初始化 第三方

	// initialize.MigrateOldData(ctx, global.GvaEntSlog, global.GvaEntDb) // 迁移旧数据

	// 打印
	slog.InfoContext(ctx, "服务启动",
		slog.Bool("debug", sys_env.GvaEnv.DEBUG),
		slog.String("version", global.Version),
		slog.String("git_commit", global.GitCommit),
		slog.String("build_time", global.BuildTime),
	)

	// 创建 Gin 服务
	ginMgr := gin_server.NewManager(global.GvaGinSlog, gin_server.Config{
		Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
		Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
		Debug:    sys_env.GvaEnv.DEBUG,
		Version:  global.Version,
	})

	// 创建 gRPC 服务
	grpcMgr := grpc_server.NewManager(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGrpcSlog,
		grpc_server.Config{
			Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
			Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
		})

	// 初始化 Gin 资源、路由
	router := ginMgr.GetServer()
	ginMgr.InitResource(ctx, embeddedFiles)

	// 配置 BasicAuth 中间件
	router.Use(gin_middleware.BasicAuthMiddleware(
		global.GvaGinSlog,
		sys_env.GvaEnv.BASIC_AUTH_USERNAME,
		sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
	))

	// 聚合所有路由注册
	ginMgr.InitRoutes()

	// 启动 MCP 服务
	mcpMgr := mcp_server.NewManager(global.GvaMcpSlog, mcp_server.Config{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.Version,
	}, router)
	go mcpMgr.Run(ctx)

	// 启动 混合 服务
	mixedMgr := mixed_manager.NewMixedServerManager(
		sys_env.GvaEnv.SERVER_ADDR,
		ginMgr,
		grpcMgr,
		global.GvaSlog,
	)
	go mixedMgr.Run(ctx)

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 启动 TUN 刷新任务
	// tunUpdateTask := tasks.NewTUNUpdateTask(
	// 	global.GvaEntDb,
	// 	global.GvaEntLogDb,
	// 	global.GvaGoCron,
	// 	global.GvaGocronSlog,
	// 	global.GvaHttpLog,
	// )
	// if err := tunUpdateTask.StartJob(ctx); err != nil {
	// 	slog.Error("启动TUN刷新任务失败",
	// 		slog.Any("err", err),
	// 	)
	// 	panic(err)
	// }

	// 启动 SSL申请任务
	// sslApplyTask := tasks.NewSSLApplyTask(
	// 	global.GvaEntDb,
	// 	global.GvaEntLogDb,
	// 	global.GvaGoCron,
	// 	global.GvaGocronSlog,
	// 	global.GvaHttpLog,
	// )
	// if err := sslApplyTask.StartPeriodicTask(ctx); err != nil {
	// 	slog.Error("启动 SSL申请任务 失败",
	// 		slog.Any("err", err),
	// 	)
	// 	panic(err)
	// }

	// 启动 SSL刷新任务
	// sslRefreshTask := tasks.NewSSLRefreshTask(
	// 	global.GvaEntDb,
	// 	global.GvaEntLogDb,
	// 	global.GvaGoCron,
	// 	global.GvaGocronSlog,
	// 	global.GvaHttpLog,
	// )
	// if err := sslRefreshTask.StartPeriodicTask(ctx); err != nil {
	// 	slog.Error("启动 SSL刷新任务 失败",
	// 		slog.Any("err", err),
	// 	)
	// 	panic(err)
	// }

	// 启动 Cloudflare隧道同步任务
	cloudflareTunnelSyncTask := tasks.NewCloudflareTunnelSyncTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := cloudflareTunnelSyncTask.StartJob(ctx); err != nil {
		slog.Error("启动 Cloudflare隧道同步任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 Cloudflare站点同步任务
	cloudflareZoneSyncTask := tasks.NewCloudflareZoneSyncTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := cloudflareZoneSyncTask.StartJob(ctx); err != nil {
		slog.Error("启动 Cloudflare站点同步任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 EdgeOne站点同步任务
	edgeOneZoneSyncTask := tasks.NewEdgeOneZoneSyncTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := edgeOneZoneSyncTask.StartJob(ctx); err != nil {
		slog.Error("启动 EdgeOne站点同步任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 ESA站点同步任务
	esaSiteSyncTask := tasks.NewESASiteSyncTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := esaSiteSyncTask.StartJob(ctx); err != nil {
		slog.Error("启动 ESA站点同步任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 清除数据任务
	clearDataTask := tasks.NewClearDataTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
	)
	if err := clearDataTask.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动 清除数据任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 创建 定时任务 管理器
	cronManager, err := cron_manager.NewCronServerManager(
		ctx,
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err != nil {
		slog.Error("创建 定时任务 管理器失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 注册任务处理器 (注入 CDN, DNS 等 Provider)
	cronManager.InitHandlers(ctx)

	// 从数据库加载并同步所有状态为 running 的任务到 gocron
	if err := cronManager.LoadAllJobs(ctx); err != nil {
		slog.Error("加载初始任务失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 将 定时任务 管理器的 SyncJob 方法注入到 Hook 包中
	schemahook.OnSyncJob = cronManager.SyncJob

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

var shutdownOnce sync.Once

// 优雅关闭实现
func gracefulShutdown() {

	shutdownOnce.Do(func() {
		slog.Info("开始优雅关闭所有服务...")

		// 创建上下文，设置超时时间为 30 秒
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 创建 Gin 服务
		ginMgr := gin_server.NewManager(global.GvaGinSlog, gin_server.Config{})

		// 创建 gRPC 服务
		grpcMgr := grpc_server.NewManager(
			global.GvaEntDb,
			global.GvaEntLogDb,
			global.GvaGrpcSlog,
			grpc_server.Config{},
		)

		// 关闭 gRPC 所有流
		if deviceService := grpcMgr.GetDeviceService(); deviceService != nil {
			deviceService.CloseAllStreams()
		}

		// 关闭 混合 服务
		mixedMgr := mixed_manager.NewMixedServerManager(
			sys_env.GvaEnv.SERVER_ADDR,
			ginMgr,
			grpcMgr,
			global.GvaSlog,
		)
		go mixedMgr.Stop(ctx)

		// 优雅关闭 MCP 服务
		mcpMgr := mcp_server.NewManager(global.GvaMcpSlog, mcp_server.Config{}, nil)
		mcpMgr.Stop(ctx)

		// 关闭 GoCron 调度器
		initialize.CloseGoCron(ctx, global.GvaGocronSlog)

		// 关闭 业务数据库
		initialize.CloseEntDb(ctx, global.GvaEntSlog)

		// 关闭 日志数据库
		initialize.CloseEntLogDb(ctx, global.GvaEntLogSlog)

		slog.Info("所有资源已优雅关闭")
	})
}
