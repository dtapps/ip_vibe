package initialize

import (
	"context"
	"log"
	"vibe_ip/internal/global"

	"github.com/go-co-op/gocron/v2"
)

// InitGoCron 初始化 GoCron 调度器
func InitGoCron(ctx context.Context) {
	s, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("[初始化 GoCron] :%v", err)
	}
	global.GvaGoCron = s
}

// CloseGoCron 关闭 GoCron 调度器
func CloseGoCron(ctx context.Context) {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.Shutdown(); err != nil {
			log.Fatalf("[关闭 GoCron] :%v", err)
		}
	}
}

// StartGoCron 启动 GoCron 调度器
func StartGoCron(ctx context.Context) {
	if global.GvaGoCron != nil {
		global.GvaGoCron.Start()
		global.GvaGoCronRunning = 1
	}
}

// StopGoCron 停止 GoCron 调度器
func StopGoCron(ctx context.Context) {
	if global.GvaGoCron != nil {
		if err := global.GvaGoCron.StopJobs(); err == nil {
			global.GvaGoCronRunning = 0
		} else {
			log.Printf("[停止 GoCron] :%v \n", err)
		}
	}
}
