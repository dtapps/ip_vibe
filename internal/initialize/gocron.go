package initialize

import (
	"context"
	"log/slog"
	"sync"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_slog"

	"github.com/go-co-op/gocron/v2"
)

var gocronMutex sync.Once

// InitGoCron 初始化 GoCron 调度器
func InitGoCron(ctx context.Context, logger *slog.Logger) {

	gocronMutex.Do(func() {

		logger.InfoContext(ctx, "GoCron 服务正在启动")

		options := make([]gocron.SchedulerOption, 0)
		// 日志
		options = append(options, gocron.WithLogger(gocron_slog.NewSlogLogger(logger)))
		// 设置时区
		options = append(options, gocron.WithLocation(time.Local))
		// 全局任务选项
		options = append(options, gocron.WithGlobalJobOptions())
		s, err := gocron.NewScheduler(options...)
		if err != nil {
			logger.ErrorContext(ctx, "[初始化 GoCron] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}
		global.GvaGoCron = s

		logger.InfoContext(ctx, "GoCron 服务启动成功")

	})
}

// CloseGoCron 关闭 GoCron 调度器
func CloseGoCron(ctx context.Context, logger *slog.Logger) {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.Shutdown(); err != nil {
			logger.Error("[关闭 GoCron] 错误",
				slog.Any("err", err),
			)
			panic(err)
		}
	}
}

// StartGoCron 启动 GoCron 调度器
func StartGoCron(ctx context.Context, logger *slog.Logger) {
	if global.GvaGoCron != nil {
		global.GvaGoCron.Start()
		global.GvaGoCronRunning = 1
	}
}

// StopGoCron 停止 GoCron 调度器
func StopGoCron(ctx context.Context, logger *slog.Logger) {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.StopJobs(); err == nil {
			global.GvaGoCronRunning = 0
		} else {
			logger.Error("[停止 GoCron] 错误",
				slog.Any("err", err),
			)
		}
	}
}
