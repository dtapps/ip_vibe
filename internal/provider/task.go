package provider

import (
	"context"
	"encoding/json"
)

// ---------- TASK 模块 ----------

// TaskProvider 定义了 定时任务 操作的接口
type TaskProvider interface {
	// Process 执行具体的任务逻辑
	// args: 对应数据库 cron_job 表中的 handler_args 通常是 JSON 字符串
	Process(ctx context.Context, args json.RawMessage) (string, error)
}
