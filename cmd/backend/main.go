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
	"vibe_ip/internal/mcp_server"
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
	global.GvaStartTime = time.Now()
}

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
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
		slog.String("os", global.GvaOS),
		slog.String("arch", global.GvaArch),
	)

	// 获取 Gin 服务
	ginMgr := gin_server.NewManager(global.GvaGinSlog, gin_server.Config{
		Addr:     sys_env.GvaEnv.HTTP_ADDR,
		Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
		Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
		Debug:    sys_env.GvaEnv.DEBUG,
		Version:  global.GvaVersion,
	})

	// 初始化 Gin 资源、路由
	router := ginMgr.GetEngine()
	ginMgr.InitResource(ctx, embeddedFiles)

	// 配置 BasicAuth 中间件
	router.Use(gin_middleware.BasicAuthMiddleware(
		global.GvaGinSlog,
		sys_env.GvaEnv.BASIC_AUTH_USERNAME,
		sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
	))

	// 聚合所有路由注册
	ginMgr.InitRoutes()

	// 启动 Gin 服务
	go ginMgr.Run(ctx)

	// 启动 MCP 服务
	mcpMgr := mcp_server.NewManager(global.GvaMcpSlog, mcp_server.Config{
		Name:    "IPVibe",
		Title:   "IPVibe",
		Version: global.GvaVersion,
	}, router)
	go mcpMgr.Run(ctx)

	// 启动 gRPC 服务
	grpcMgr := grpc_server.NewManager(global.GvaGrpcSlog, grpc_server.Config{
		Addr:     sys_env.GvaEnv.GRPC_ADDR,
		Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
		Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
	})
	go grpcMgr.Run(ctx)

	// 连接 gRPC-Gateway 服务
	// grpcGwMgr := grpc_gateway_client.NewManager(global.GvaGrpcSlog, grpc_gateway_client.Config{
	// 	Addr:     sys_env.GvaEnv.GRPC_ADDR,
	// 	Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
	// 	Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
	// }, router)
	// go grpcGwMgr.Conn(ctx)

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 启动 CDN/DNS刷新任务
	ipUpdateTask := tasks.NewIPUpdateTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := ipUpdateTask.StartCdnJob(ctx); err != nil {
		slog.Error("启动 CDN/DNS刷新任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

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
	sslApplyTask := tasks.NewSSLApplyTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := sslApplyTask.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动 SSL申请任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 SSL刷新任务
	sslRefreshTask := tasks.NewSSLRefreshTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := sslRefreshTask.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动 SSL刷新任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 启动 域名监控任务
	domainMonitorTask := tasks.NewDomainMonitorTask(
		global.GvaEntDb,
		global.GvaEntLogDb,
		global.GvaGoCron,
		global.GvaGocronSlog,
		global.GvaHttpLog,
	)
	if err := domainMonitorTask.StartPeriodicTask(ctx); err != nil {
		slog.Error("启动 域名监控任务 失败",
			slog.Any("err", err),
		)
		panic(err)
	}

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

		// 优雅关闭 Gin 服务
		ginMgr := gin_server.NewManager(global.GvaGinSlog, gin_server.Config{})
		ginMgr.Stop(ctx)

		// 优雅关闭 MCP 服务
		mcpMgr := mcp_server.NewManager(global.GvaMcpSlog, mcp_server.Config{}, nil)
		mcpMgr.Stop(ctx)

		// 优雅关闭 gRPC-Gateway 服务
		// grpc_gateway_client.NewManager(global.GvaGrpcSlog, grpc_gateway_client.Config{}, nil).Stop(ctx)

		// 优雅关闭 gRPC 服务
		grpcMgr := grpc_server.NewManager(global.GvaGrpcSlog, grpc_server.Config{})
		if agentSvc := grpcMgr.GetAgentService(); agentSvc != nil {
			// 主动关闭所有流
			agentSvc.CloseAllStreams()
		}
		grpcMgr.Stop(ctx)

		// 6. 关闭数据库连接和其他资源（
		slog.Info("开始关闭数据库连接...")

		// 7. 关闭 GoCron 调度器
		initialize.CloseGoCron(ctx, global.GvaGocronSlog)

		// 8. 等待一段时间让 goroutine 完成
		time.Sleep(1 * time.Second)

		// 9. 关闭 业务数据库
		initialize.CloseEntDb(ctx, global.GvaEntSlog)

		// 10. 关闭 日志数据库
		initialize.CloseEntLogDb(ctx, global.GvaEntLogSlog)

		slog.Info("所有资源已优雅关闭")
	})
}
