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
	"vibe_ip/internal/grpc_credential"
	"vibe_ip/internal/initialize"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
				slog.String("error", errMsg),
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
	initialize.InitRestySLog("forwarder")  // 初始化 resty 日志
	initialize.InitHttpSLog("forwarder")   // 初始化 http 日志
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
		// 获取 gRPC 客户端
		grpcClient, err := get_grpc_client(ctx, global.GvaGrpcSlog)
		if err != nil {
			fmt.Printf("获取 gRPC 客户端失败: %v\n", err)
			panic(err)
		}
		defer grpcClient.Close()

		// 创建端口转发实例
		portForwardTask := task.NewPortForwardGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			grpcClient,
		)
		portForwardTask.SetConfig(config)
		// 启动端口转发任务
		err = portForwardTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动端口转发任务失败: %v\n", err)
			panic(err)
		}
		// 关闭端口转发实例
		defer portForwardTask.Close()

		// 创建统计上报实例
		statsReportTask := task.NewStatsReportGRPCTask(
			global.GvaGoCron,
			global.GvaGocronSlog,
			portForwardServer,
			forwardAgent,
			grpcClient,
		)
		statsReportTask.SetConfig(config)
		// 启动统计上报任务
		err = statsReportTask.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动统计上报任务失败: %v\n", err)
			panic(err)
		}
		// 关闭统计上报实例
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
			fmt.Printf("启动端口转发任务失败: %v\n", err)
			panic(err)
		}
		// 关闭端口转发实例
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
			fmt.Printf("启动统计上报任务失败: %v\n", err)
			panic(err)
		}
		// 关闭统计上报实例
		defer statsReportTask.Close()
	} else {
		fmt.Printf("服务启动失败: GRPC_URL / HTTP_URL未配置\n")
		panic("GRPC_URL / HTTP_URL未配置")
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 阻止主线程退出
	select {}
}

// get_grpc_client 获取 gRPC 客户端
func get_grpc_client(ctx context.Context, grpcLogger *slog.Logger) (*grpc.ClientConn, error) {
	// 创建 gRPC 客户端
	grpc_client := grpc_client.NewGRPCClient()
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	// 添加 gRPC 连接选项
	grpc_client.AddDialOption(
		// 鉴权
		grpc.WithPerRPCCredentials(grpc_credential.NewClientBasicAuth(
			grpcLogger,
			global.GvaEnv.BASIC_AUTH_USERNAME,
			global.GvaEnv.BASIC_AUTH_PASSWORD,
		)),
	)
	// 添加 retry 中间件（Unary 和 Stream）
	grpc_client.AddDialOption(
		grpc.WithUnaryInterceptor(retry.UnaryClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
		)),
	)
	grpc_client.AddDialOption(
		grpc.WithStreamInterceptor(retry.StreamClientInterceptor(
			retry.WithMax(5), // 最多 5 次尝试（初始 + 4 重试）
			retry.WithBackoff(retry.BackoffExponential(100*time.Millisecond)), // 指数退避，初始 100ms
			retry.WithCodes(codes.Unavailable),                                // 只针对 Unavailable 重试（你的 connection refused）
			retry.WithPerRetryTimeout(5*time.Second),                          // 每个重试超时 5s
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
