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

				// 如果 OnSyncJob 为空，直接返回
				if OnSyncJob == nil {
					return next.Mutate(ctx, m)
				}

				// 获取 ID (变更前抓取)
				var objIDS []int64
				if m.Op().Is(ent.OpUpdate | ent.OpDelete) {
					if ids, err := m.IDs(ctx); err == nil && len(ids) > 0 {
						objIDS = ids
					}
				} else if id, exists := m.ID(); exists {
					objIDS = []int64{id}
				}

				// 执行业务变更
				result, err := next.Mutate(ctx, m)
				if err != nil {
					return result, err
				}

				// 补偿获取 Create 后的 ID
				if m.Op().Is(ent.OpCreate) && result != nil {
					if id, ok := m.ID(); ok {
						objIDS = []int64{id}
					}
				}

				// 异步同步任务状态
				if len(objIDS) > 0 {
					go func(asyncCtx context.Context, asyncIDs []int64) {
						for _, asyncID := range asyncIDs {
							// 使用独立的 context 保证执行
							_ = OnSyncJob(asyncCtx, asyncID)
						}
					}(context.WithoutCancel(ctx), objIDS)
				}

				return result, nil
			})
		},
		// ent.OpCreate,
		ent.OpCreate|ent.OpDelete|ent.OpDeleteOne,
	)
}
