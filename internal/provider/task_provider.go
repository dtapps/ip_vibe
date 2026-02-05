package provider

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
)

// ---------- TASK 模块 ----------

// TaskProvider 定义了 定时任务 操作的接口
type TaskProvider interface {
	// Process 执行具体的任务逻辑
	// args: 对应数据库 cron_job 表中的 handler_args 通常是 JSON 字符串
	Process(ctx context.Context, args json.RawMessage) (string, error)
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
