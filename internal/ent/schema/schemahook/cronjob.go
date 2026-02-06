package schemahook

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/hook"
)

// 定义一个同步函数的类型，避免引用 cron_manager
type SyncJobFunc func(ctx context.Context, id int64) error

// 全局变量，用于存储实际的同步逻辑
var OnSyncJob SyncJobFunc

// CronJobLogHook 为 CronJob 实现创建时就添加到定时任务调度器
func CronJobGocronHook() ent.Hook {
	return hook.On(
		func(next ent.Mutator) ent.Mutator {
			return hook.CronJobFunc(func(ctx context.Context, m *ent.CronJobMutation) (ent.Value, error) {
				// 先执行数据库变更（获取执行结果）
				value, err := next.Mutate(ctx, m)
				if err != nil {
					return value, err
				}

				// 判断 OnSyncJob 是否已注入
				if OnSyncJob == nil {
					return value, nil
				}

				// 提取任务 ID
				id, exists := m.ID()
				if !exists {
					// 如果是 OpCreate，ID 存放在返回的实体中
					if node, ok := value.(*ent.CronJob); ok {
						id = node.ID
						exists = true
					}
				}

				// 4. 异步同步任务状态
				if exists {
					go func(asyncCtx context.Context, asyncID int64) {
						// SyncJob 内部已处理：查不到 ID = 移除，状态不符 = 移除，配置变动 = 重启
						_ = OnSyncJob(asyncCtx, asyncID)
					}(context.WithoutCancel(ctx), id)
				}

				return value, nil
			})
		},
		ent.OpCreate,
	)
}
