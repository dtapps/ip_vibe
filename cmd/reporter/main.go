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
	"vibe_ip/internal/grpc_generated"
	"vibe_ip/internal/helper/sys_env"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/reporter/service"
	"vibe_ip/internal/reporter/tasks"

	"github.com/denisbrodbeck/machineid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	version   = "dev"     // 版本号
	commit    = "unknown" // Git 提交 ID
	buildTime = "unknown" // 构建时间
)

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
	initialize.InitGocronSLog("reporter") // 初始化 gocron 日志
	initialize.InitHttpSLog("reporter")   // 初始化 http 日志
	initialize.InitRestySLog("reporter")  // 初始化 resty 日志
	initialize.EnsureDefaultSLog()        // 确保日志不为 nil

	initialize.InitGoCron(ctx, global.GvaGocronSlog) // 初始化 GoCron 调度器
	defer func() {
		initialize.CloseGoCron(ctx, global.GvaGocronSlog) // 关闭 GoCron 调度器
	}()

	// 打印
	slog.Info("服务启动",
		slog.Bool("debug", sys_env.GvaEnv.DEBUG),
		slog.String("version", global.Version),
		slog.String("git_commit", global.GitCommit),
		slog.String("build_time", global.BuildTime),
	)

	// 设备编码
	if sys_env.GvaEnv.DEVICE_CODE == "" {
		sys_env.GvaEnv.DEVICE_CODE, _ = machineid.ID()
	}

	if sys_env.GvaEnv.BACKEND_ENDPOINT != "" {

		// 连接 gRPC 客户端
		grpcMgr := grpc_client.NewManager(global.GvaGrpcSlog, grpc_client.Config{
			Addr:     sys_env.GvaEnv.BACKEND_ENDPOINT,
			Username: sys_env.GvaEnv.BASIC_AUTH_USERNAME,
			Password: sys_env.GvaEnv.BASIC_AUTH_PASSWORD,
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
		ipReportTask, err := tasks.NewIPReportGRPCTask(
			ctx,
			global.GvaGoCron,
			global.GvaGocronSlog,
			grpcClientConn,
		)
		if err != nil {
			slog.ErrorContext(ctx, "创建IP上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}

		// 启动IP上报任务
		err = ipReportTask.StartPeriodicTask(ctx)
		if err != nil {
			slog.ErrorContext(ctx, "启动IP上报任务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		defer ipReportTask.Close()

		// 设备服务 启动流式通信
		deviceService, err := service.NewDeviceService(ctx, global.GvaGocronSlog, grpcClientConn)
		if err != nil {
			slog.ErrorContext(ctx, "创建设备服务失败",
				slog.Any("err", err),
			)
			panic(err)
		}
		deviceService.SetClientInfo(&grpc_generated.ClientInfo{
			Version:   global.Version,
			GitCommit: global.GitCommit,
			BuildTime: global.BuildTime,
			RunTime:   timestamppb.New(global.StartTime),
		})
		go deviceService.Run(ctx)

	} else {
		slog.ErrorContext(ctx, "服务启动失败: BACKEND_ENDPOINT 未配置")
		panic("服务启动失败: BACKEND_ENDPOINT 未配置")
	}

	// 启动 GoCron 调度器
	initialize.StartGoCron(ctx, global.GvaGocronSlog)

	// 阻止主线程退出
	select {}
}
