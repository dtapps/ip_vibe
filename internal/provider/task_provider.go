package provider

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"vibe_ip/internal/ent/cronjob"
)

// ---------- TASK 模块 ----------

// TaskMetadata 任务元数据：描述执行时的背景信息
type TaskMetadata struct {
	JobID        int64               // 任务序号
	JobName      string              // 任务名称
	TaskType     cronjob.TaskType    // 任务类型: one_off 一次性, recurring 周期性
	TaskCategory cronjob.JobCategory // 任务分类: system 系统内置, custom 用户自定义
}

// TaskProvider 定义了 定时任务 操作的接口
type TaskProvider interface {
	// Process 执行具体的任务逻辑
	// args: 对应数据库 cron_job 表中的 handler_args 通常是 JSON 字符串
	Process(ctx context.Context, args json.RawMessage, meta TaskMetadata) (string, error)
}

// TaskTask 定义了 定时任务
type TaskTask struct {
	logger   *slog.Logger            // 日志记录器
	mu       sync.RWMutex            // 引入读写锁
	handlers map[string]TaskProvider // 任务提供商映射
}

// NewTaskTask 创建一个新的定时任务实例
func NewTaskTask(ctx context.Context, logger *slog.Logger) (t *TaskTask, err error) {
	t = &TaskTask{
		logger:   logger,
		handlers: make(map[string]TaskProvider),
	}

	// 优化日志
	t.logger = t.logger.With(
		slog.String("模块", "Task"),
	)

	return t, nil
}

// RegisterTask 注册具体的任务处理器
func (t *TaskTask) RegisterTask(ctx context.Context, name string, handler TaskProvider) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.handlers[name] = handler
	t.logger.InfoContext(ctx, "注册任务处理器",
		slog.String("handler_name", name),
	)
}

// GetTask 获取处理器
func (t *TaskTask) GetTask(name string) (TaskProvider, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	h, ok := t.handlers[name]
	return h, ok
}
