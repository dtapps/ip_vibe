package initialize

import (
	"log"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/gocron_slog"

	"github.com/go-co-op/gocron/v2"
)

// InitGoCron 初始化 GoCron 调度器
func InitGoCron() {
	options := make([]gocron.SchedulerOption, 0)
	if global.GvaGocronSlog != nil {
		options = append(options, gocron.WithLogger(gocron_slog.NewSlogLogger(global.GvaGocronSlog.WithLogger())))
	}
	s, err := gocron.NewScheduler(options...)
	if err != nil {
		log.Fatalf("[初始化 GoCron] :%v", err)
	}
	global.GvaGoCron = s
}

// CloseGoCron 关闭 GoCron 调度器
func CloseGoCron() {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.Shutdown(); err != nil {
			log.Fatalf("[关闭 GoCron] :%v", err)
		}
	}
}

// StartGoCron 启动 GoCron 调度器
func StartGoCron() {
	if global.GvaGoCron != nil {
		global.GvaGoCron.Start()
		global.GvaGoCronRunning = 1
	}
}

// StopGoCron 停止 GoCron 调度器
func StopGoCron() {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.StopJobs(); err == nil {
			global.GvaGoCronRunning = 0
		} else {
			log.Printf("[停止 GoCron] :%v \n", err)
		}
	}
}
