package schemahook

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/global"

	"github.com/spf13/cast"
)

// getEntSlog: 获取业务数据库日志实例
func getEntSlog() *slog.Logger {
	logger := slog.Default()
	if global.GvaEntSlog != nil {
		logger = global.GvaEntSlog
	}
	return logger
}

// getEntLogSlog: 获取操作日志数据库日志实例
func getEntLogSlog() *slog.Logger {
	logger := slog.Default()
	if global.GvaEntLogSlog != nil {
		logger = global.GvaEntLogSlog
	}
	return logger
}

// getDb: 获取业务数据库客户端
func getDb() (db *ent.Client) {
	return global.GvaEntDb
}

// getLogDb: 获取操作日志数据库客户端
func getLogDb() (db *ent_log.Client) {
	return global.GvaEntLogDb
}

// getObjID: 从 ent.Mutation 或 ent.Value 中获取对象 ID
func getObjID[M ent.Mutation](ctx context.Context, op ent.Op, m M, result ent.Value) (string, error) {

	getEntSlog().DebugContext(ctx, "从 ent.Mutation 或 ent.Value 中获取对象 ID",
		slog.String("op", op.String()),
		slog.Any("m", m),
		slog.Any("result", result),
	)

	// 1. 处理 Create 场景 (从 result 获取)
	if op.Is(ent.OpCreate) && result != nil {
		// 先转为 any 再断言
		if v, ok := any(result).(interface{ GetID() int64 }); ok {
			return cast.ToString(v.GetID()), nil
		}
	}

	// 2. 处理单体操作 UpdateOne / DeleteOne (直接从 Mutation 获取)
	if op.Is(ent.OpUpdateOne | ent.OpDeleteOne) {
		if v, ok := any(m).(interface{ ID() (int64, bool) }); ok {
			if val, exists := v.ID(); exists {
				return cast.ToString(val), nil
			}
		}
	}

	// 3. 处理批量操作 Update / Delete (通过条件反查所有 ID)
	if op.Is(ent.OpUpdate | ent.OpDelete) {
		if v, ok := any(m).(interface {
			IDs(ctx context.Context) ([]int64, error)
		}); ok {
			if vals, err := v.IDs(ctx); err == nil {
				return strings.Join(cast.ToStringSlice(vals), ","), nil
			}
		}
	}

	return "", fmt.Errorf("无法获取对象 ID, Op: %s", op.String())
}

// getChangedFields: 返回 JSON 字符串以适配你的 Schema
func getChangedFields[M ent.Mutation](ctx context.Context, m M) (map[string]any, map[string]any) {
	changed := m.Fields()
	oldMap := make(map[string]any)
	newMap := make(map[string]any)

	// 获取新值
	for _, f := range changed {
		if val, exists := m.Field(f); exists {
			newMap[f] = val
		}
	}

	// 获取旧值 (关键点：使用 OldField 前确保断言成功)
	if v, ok := any(m).(interface {
		OldField(context.Context, string) (any, error)
	}); ok {
		for _, f := range changed {
			// 注意：OldField 可能会因为数据不存在或未加载而报错
			if val, err := v.OldField(ctx, f); err == nil {
				oldMap[f] = val
			}
		}
	}

	return oldMap, newMap
}
