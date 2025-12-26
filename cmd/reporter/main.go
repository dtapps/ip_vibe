package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"vibe_ip/internal/global"
	"vibe_ip/internal/grpc_client"
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/reporter/service"
	"vibe_ip/internal/reporter/task"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	initialize.InitSLog("reporter")       // 初始化 日志
	initialize.InitGrpcSLog("reporter")   // 初始化 gRPC 日志
	initialize.InitGocronSLog("reporter") // 初始化 gocron 日志

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
	ipReportService := service.NewIPReportService(global.GvaGocronSlog.WithLogger())

	if global.GvaEnv.GRPC_URL != "" {
		// 获取 gRPC 客户端
		grpcClient, err := get_grpc_client()
		if err != nil {
			fmt.Printf("获取 gRPC 客户端失败: %v\n", err)
			os.Exit(1)
		}
		defer grpcClient.Close()

		// 创建IP上报实例
		ipReportTask := task.NewIPReportGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog.WithLogger(),
			ipReportService,
			grpcClient,
		)
		ipReportTask.SetConfig(config)
		// 启动IP上报任务
		err = ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动任务失败: %v\n", err)
			os.Exit(1)
		}
		// 关闭IP上报实例
		defer ipReportTask.Close()
	} else if global.GvaEnv.HTTP_URL != "" {
		// 创建IP上报实例
		ipReportTask := task.NewIPReportRestyTask(
			global.GvaGoCron,
			global.GvaGocronSlog.WithLogger(),
			ipReportService,
		)
		ipReportTask.SetDebug(global.GvaEnv.DEBUG)
		ipReportTask.SetConfig(config)
		ipReportTask.SetBaseURL(global.GvaEnv.HTTP_URL)
		ipReportTask.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		// 启动IP上报任务
		err := ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动任务失败: %v\n", err)
			os.Exit(1)
		}
		// 关闭IP上报实例
		defer ipReportTask.Close()
	} else {
		fmt.Printf("服务启动失败: GRPC_URL / HTTP_URL未配置\n")
		os.Exit(1)
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron()

	// 阻止主线程退出
	select {}
}

// get_grpc_client 获取 gRPC 客户端
func get_grpc_client() (*grpc.ClientConn, error) {
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
	err := grpc_client.Dial(global.GvaEnv.GRPC_URL)
	if err != nil {
		return nil, fmt.Errorf("连接gRPC服务器失败: %w", err)
	}
	// 获取 gRPC 客户端
	return grpc_client.GetClient(), nil
}
