package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/forwarder/task"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_client"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/thrift_client"
)

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

	ctx := context.Background()

	initialize.InitSLog("forwarder")       // 初始化 日志
	initialize.InitGrpcSLog("forwarder")   // 初始化 gRPC 日志
	initialize.InitThriftSLog("forwarder") // 初始化 Thrift 日志
	initialize.InitGocronSLog("forwarder") // 初始化 gocron 日志
	initialize.InitHttpSLog("forwarder")   // 初始化 http 日志
	initialize.InitRestySLog("forwarder")  // 初始化 resty 日志
	initialize.InitReqSLog("forwarder")    // 初始化 req 日志
	initialize.EnsureDefaultSLog()         // 确保日志不为 nil

	initialize.InitGoCron(ctx, global.GvaGocronSlog) // 初始化 GoCron 调度器
	defer func() {
		initialize.CloseGoCron(ctx, global.GvaGocronSlog) // 关闭 GoCron 调度器
	}()

	// 打印
	slog.InfoContext(ctx, "服务启动",
		slog.Bool("debug", global.GvaEnv.DEBUG),
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
		slog.String("os", global.GvaOS),
		slog.String("arch", global.GvaArch),
	)

	// 配置
	config := &task.PortForwardTaskConfig{
		DeviceCode: global.GvaEnv.DEVICE_CODE,
		DeviceName: global.GvaEnv.DEVICE_NAME,
		Version:    global.GvaVersion,
		GitCommit:  global.GvaGitCommit,
		BuildTime:  global.GvaBuildTime,
		GoVersion:  global.GvaGoVersion,
		OS:         global.GvaOS,
		Arch:       global.GvaArch,
	}

	// 创建端口转发服务
	portForwardServer := service.NewPortForwardServer(global.GvaGocronSlog)

	// 创建端口转发器
	forwardAgent := forward.NewAgent(global.GvaGocronSlog)

	if global.GvaEnv.GRPC_URL != "" {

		// 连接 gRPC 客户端
		grpcMgr := grpc_client.NewManager(global.GvaGrpcSlog, grpc_client.Config{
			Addr:     global.GvaEnv.GRPC_URL,
			Username: global.GvaEnv.BASIC_AUTH_USERNAME,
			Password: global.GvaEnv.BASIC_AUTH_PASSWORD,
		})
		err := grpcMgr.Conn(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动 gRPC 客户端失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer grpc_client.NewManager(global.GvaGrpcSlog, grpc_client.Config{}).Stop(ctx)

		// 获取 gRPC 客户端连接
		grpcClientConn := grpcMgr.GetConn()

		// 创建端口转发实例
		portForwardTask := task.NewPortForwardGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			grpcClientConn,
		)
		portForwardTask.SetConfig(config)

		// 启动端口转发任务
		err = portForwardTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动端口转发任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer portForwardTask.Close()

		// 创建统计上报实例
		statsReportTask := task.NewStatsReportGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			grpcClientConn,
		)
		statsReportTask.SetConfig(config)

		// 启动统计上报任务
		err = statsReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动统计上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer statsReportTask.Close()

		// 启动远程日志隧道
		remoteLogService := service.NewRemoteLogService(global.GvaGocronSlog, grpcClientConn)
		go remoteLogService.StartTunnel(ctx, config.DeviceCode)

	} else if global.GvaEnv.THRIFT_URL != "" {

		// 连接 Thrift 客户端
		thriftMgr := thrift_client.NewManager(global.GvaThriftSlog, thrift_client.Config{
			Addr:     global.GvaEnv.THRIFT_URL,
			Username: global.GvaEnv.BASIC_AUTH_USERNAME,
			Password: global.GvaEnv.BASIC_AUTH_PASSWORD,
		})
		err := thriftMgr.Conn(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动 Thrift 客户端失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer thrift_client.NewManager(global.GvaThriftSlog, thrift_client.Config{}).Stop(ctx)

		// 创建端口转发实例
		portForwardTask := task.NewPortForwardThriftTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			thriftMgr.GetConn(),
			thriftMgr.GetAuth(),
		)
		portForwardTask.SetConfig(config)

		// 启动端口转发任务
		err = portForwardTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动端口转发任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer portForwardTask.Close()

		// 创建统计上报实例
		statsReportTask := task.NewStatsReportThriftTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			thriftMgr.GetConn(),
			thriftMgr.GetAuth(),
		)
		statsReportTask.SetConfig(config)

		// 启动统计上报任务
		err = statsReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动统计上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer statsReportTask.Close()

	} else if global.GvaEnv.HTTP_URL != "" {

		// 创建端口转发实例
		portForwardTask := task.NewPortForwardRestyTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
		)
		portForwardTask.SetDebug(global.GvaEnv.DEBUG)
		portForwardTask.SetConfig(config)
		portForwardTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		portForwardTask.SetBaseURL(global.GvaEnv.HTTP_URL)

		// 启动端口转发任务
		err := portForwardTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动端口转发任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer portForwardTask.Close()

		// 创建统计上报实例
		statsReportTask := task.NewStatsReportRestyTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
		)
		statsReportTask.SetConfig(config)
		statsReportTask.SetBaseURL(global.GvaEnv.HTTP_URL)
		statsReportTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)

		// 启动统计上报任务
		err = statsReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动统计上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer statsReportTask.Close()

	} else {
		slog.ErrorContext(ctx, "服务启动失败: GRPC_URL / THRIFT_URL / HTTP_URL未配置")
		panic("GRPC_URL / THRIFT_URL / HTTP_URL未配置")
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 阻止主线程退出
	select {}
}
