package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
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

	if global.GvaEnv.GRPC_URL != "" {
	} else if global.GvaEnv.HTTP_URL != "" {
		// 创建服务
		service := task.NewPortForwardRestyTask(global.GvaGoCron)
		service.SetDebug(global.GvaEnv.DEBUG)
		service.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
		service.SetBaseURL(global.GvaEnv.HTTP_URL)
		// 启动任务
		err := service.StartPeriodicTask(ctx)
		if err != nil {
			fmt.Printf("启动任务失败: %v\n", err)
			os.Exit(1)
		}
		// 关闭服务
		defer service.Close()
	} else {
		fmt.Printf("服务启动失败: GRPC_URL / HTTP_URL未配置\n")
		os.Exit(1)
	}

	// 阻止主线程退出
	select {}
}
