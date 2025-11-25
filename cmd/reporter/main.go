package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"vibe_ip/internal/global"
	"vibe_ip/internal/initialize"
	"vibe_ip/internal/reporter"
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

	initialize.InitGoCron(ctx)           // 初始化 GoCron
	initialize.InitSLog(ctx, "reporter") // 初始化 日志
	defer func() {
		initialize.CloseGoCron(ctx) // 关闭 GoCron
	}()

	// 打印
	slog.Info("服务启动",
		slog.Bool("debug", global.GvaEnv.DEBUG),
		slog.String("version", global.GvaVersion),
		slog.String("git_commit", global.GvaGitCommit),
		slog.String("build_time", global.GvaBuildTime),
		slog.String("go_version", global.GvaGoVersion),
	)

	if global.GvaEnv.HTTP_URL == "" {
		fmt.Printf("服务启动失败: HTTP_URL未配置\n")
		os.Exit(1)
	}

	// 规范化HTTP_URL并确保以/api/devices/ip-report结尾
	// 首先移除可能存在的重复api/devices/ip-report或/api/devices/ip-report后缀
	targetPath := "/api/devices/ip-report"
	altPath := "api/devices/ip-report"

	// 规范化URL，移除末尾可能的斜杠
	url := strings.TrimSuffix(global.GvaEnv.HTTP_URL, "/")

	// 检查并移除已存在的路径后缀
	if strings.HasSuffix(url, targetPath) {
		url = strings.TrimSuffix(url, targetPath)
	} else if strings.HasSuffix(url, altPath) {
		url = strings.TrimSuffix(url, altPath)
	}

	// 添加正确格式的路径
	global.GvaEnv.HTTP_URL = url + targetPath

	// 创建服务
	service := reporter.NewService(global.GvaGoCron)
	if global.GvaEnv.DEBUG {
		service.EnableDebug()
	}
	if global.GvaEnv.BASIC_AUTH_USERNAME != "" && global.GvaEnv.BASIC_AUTH_PASSWORD != "" {
		service.SetBasicAuth(global.GvaEnv.BASIC_AUTH_USERNAME, global.GvaEnv.BASIC_AUTH_PASSWORD)
	}
	service.SetBaseURL(global.GvaEnv.HTTP_URL)

	// 启动任务
	err := service.StartJob(ctx)
	if err != nil {
		fmt.Printf("启动任务失败: %v\n", err)
		os.Exit(1)
	}

	// 启动 Gocron
	initialize.StartGoCron(ctx)

	// 阻止主线程退出
	select {}
}
