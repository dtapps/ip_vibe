package global

import (
	"vibe_ip/ent"
	"vibe_ip/internal/helper/operation_logs_helper"

	"github.com/go-co-op/gocron/v2"
)

// GvaConfig 变量
var (
	GvaEnv struct {
		DEBUG               bool   // [backend][reporter] 是否开启调试模式
		HTTP_PORT           string // [backend] HTTP 端口
		HTTP_URL            string // [reporter] HTTP URL
		BASIC_AUTH_USERNAME string // [backend][reporter] 基本认证用户名
		BASIC_AUTH_PASSWORD string // [backend][reporter] 基本认证密码
		DEVICE_CODE         string // [reporter] 设备代码
		DEVICE_NAME         string // [reporter] 设备名称
	}

	GvaGoCron        gocron.Scheduler // 定时任务
	GvaGoCronRunning int32            // 定时任务状态 0=停 1=运行

	GvaDb *ent.Client // 数据库

	GvaOperationLogHelper *operation_logs_helper.OperationLogHelper // 操作日志助手
)
