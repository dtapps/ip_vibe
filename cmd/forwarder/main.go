package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"vibe_ip/internal/forwarder/forward"
	"vibe_ip/internal/forwarder/service"
	"vibe_ip/internal/forwarder/task"
	"vibe_ip/internal/global"
	"vibe_ip/internal/initialize"
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

	ctx := context.Background()

	initialize.InitSLog("forwarder")     // 初始化 日志
	initialize.InitGocronSLog("backend") // 初始化 gocron 日志

	initialize.InitGoCron() // 初始化 GoCron 调度器
	defer func() {
		initialize.CloseGoCron() // 关闭 GoCron 调度器
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
	portForwardServer := service.NewPortForwardServer(global.GvaGocronSlog.WithLogger())

	// 创建端口转发器
	forwardAgent := forward.NewAgent(global.GvaGocronSlog.WithLogger())

	if global.GvaEnv.GRPC_URL != "" {
		// 创建服务
		grpcTask := task.NewPortForwardGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog.WithLogger(),
			portForwardServer,
			forwardAgent,
		)
		grpcTask.SetConfig(config)
		grpcTask.SetBaseURL(global.GvaEnv.GRPC_URL)
		grpcTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		// 启动任务
		err := grpcTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动任务失败: %v\n", err)
			os.Exit(1)
		}
		// 关闭服务
		defer grpcTask.Close()
	} else if global.GvaEnv.HTTP_URL != "" {
		// 创建服务
		restyTask := task.NewPortForwardRestyTask(
			global.GvaGoCron,
			global.GvaGocronSlog.WithLogger(),
			portForwardServer,
			forwardAgent,
		)
		restyTask.SetDebug(global.GvaEnv.DEBUG)
		restyTask.SetConfig(config)
		restyTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		restyTask.SetBaseURL(global.GvaEnv.HTTP_URL)
		// 启动任务
		err := restyTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动任务失败: %v\n", err)
			os.Exit(1)
		}
		// 关闭服务
		defer restyTask.Close()
	} else {
		fmt.Printf("服务启动失败: GRPC_URL / HTTP_URL未配置\n")
		os.Exit(1)
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 阻止主线程退出
	select {}
}
