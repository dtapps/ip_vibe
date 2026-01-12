package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_client"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/reporter/service"
	"vibe_ip/internal/reporter/task"
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

	initialize.InitSLog("reporter")       // 初始化 日志
	initialize.InitGrpcSLog("reporter")   // 初始化 gRPC 日志
	initialize.InitThriftSLog("reporter") // 初始化 Thrift 日志
	initialize.InitGocronSLog("reporter") // 初始化 gocron 日志
	initialize.InitHttpSLog("reporter")   // 初始化 http 日志
	initialize.InitRestySLog("reporter")  // 初始化 resty 日志
	initialize.InitReqSLog("reporter")    // 初始化 req 日志
	initialize.EnsureDefaultSLog()        // 确保日志不为 nil

	initialize.InitGoCron(ctx, global.GvaGocronSlog) // 初始化 GoCron 调度器
	defer func() {
		initialize.CloseGoCron(ctx, global.GvaGocronSlog) // 关闭 GoCron 调度器
	}()

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

	// 配置
	config := &task.IPReportTaskConfig{
		DeviceCode: global.GvaEnv.DEVICE_CODE,
		DeviceName: global.GvaEnv.DEVICE_NAME,
		Version:    global.GvaVersion,
		GitCommit:  global.GvaGitCommit,
		BuildTime:  global.GvaBuildTime,
		GoVersion:  global.GvaGoVersion,
		OS:         global.GvaOS,
		Arch:       global.GvaArch,
	}

	// 创建 IP 报告服务
	ipReportService := service.NewIPReportService(global.GvaGocronSlog)

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

		// 创建IP上报实例
		ipReportTask := task.NewIPReportGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			ipReportService,
			grpcClientConn,
		)
		ipReportTask.SetConfig(config)

		// 启动IP上报任务
		err = ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动IP上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer ipReportTask.Close()

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

		// 创建IP上报实例
		ipReportTask := task.NewIPReportThriftTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			ipReportService,
			thriftMgr.GetConn(),
			thriftMgr.GetAuth(),
		)
		ipReportTask.SetConfig(config)

		// 启动IP上报任务
		err = ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动IP上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer ipReportTask.Close()

	} else if global.GvaEnv.HTTP_URL != "" {

		// 创建IP上报实例
		ipReportTask := task.NewIPReportRestyTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			ipReportService,
		)
		ipReportTask.SetDebug(global.GvaEnv.DEBUG)
		ipReportTask.SetConfig(config)
		ipReportTask.SetBaseURL(global.GvaEnv.HTTP_URL)
		ipReportTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)

		// 启动IP上报任务
		err := ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动IP上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer ipReportTask.Close()

	} else {
		slog.ErrorContext(ctx, "服务启动失败: GRPC_URL / THRIFT_URL / HTTP_URL未配置")
		panic("GRPC_URL / THRIFT_URL / HTTP_URL未配置")
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 阻止主线程退出
	select {}
}
