package schemahook

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/hook"
	"vibe_ip/internal/ent/platformconfig"
)

// 定义实际实现类型，避免引用 cron_manager
type PlatformConfigGocronHookFunc func(ctx context.Context, record *ent.PlatformConfig) error

// 全局变量，用于存储实际逻辑
var OnPlatformConfigGocronHook PlatformConfigGocronHookFunc

// PlatformConfigGocronHook 为 PlatformConfig 添加数据定时任务钩子
func PlatformConfigGocronHook() ent.Hook {
	return hook.On(
		func(next ent.Mutator) ent.Mutator {
			return hook.PlatformConfigFunc(func(ctx context.Context, m *ent.PlatformConfigMutation) (ent.Value, error) {

				// 如果 实现 为空，直接返回
				if OnPlatformConfigGocronHook == nil {
					return next.Mutate(ctx, m)
				}

				// 获取 记录 (变更前抓取)
				var objRecords []*ent.PlatformConfig
				if m.Op().Is(ent.OpUpdate | ent.OpDelete) {
					if ids, err := m.IDs(ctx); err == nil && len(ids) > 0 {
						if records, err := m.Client().PlatformConfig.Query().Where(platformconfig.IDIn(ids...)).All(ctx); err == nil {
							objRecords = records
						}
					}
				} else if id, exists := m.ID(); exists {
					if record, err := m.Client().PlatformConfig.Get(ctx, id); err == nil {
						objRecords = []*ent.PlatformConfig{record}
					}
				}

				// 执行业务变更
				result, err := next.Mutate(ctx, m)
				if err != nil {
					return result, err
				}

				// 补偿获取 Create 后的 ID
				if m.Op().Is(ent.OpCreate) && result != nil {
					if id, ok := m.ID(); ok {
						if record, err := m.Client().PlatformConfig.Get(ctx, id); err == nil {
							objRecords = []*ent.PlatformConfig{record}
						}
					}
				}

				// 异步同步任务状态
				if len(objRecords) > 0 {
					go func(asyncCtx context.Context, asyncRecords []*ent.PlatformConfig) {
						for _, asyncRecord := range asyncRecords {
							// 使用独立的 context 保证执行
							_ = OnPlatformConfigGocronHook(asyncCtx, asyncRecord)
						}
					}(context.WithoutCancel(ctx), objRecords)
				}

				return result, nil
			})
		},
		ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
	)
}
