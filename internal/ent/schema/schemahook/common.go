package schemahook

import (
	"context"
	"fmt"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
)

var entSlog = global.GvaEntSlog.WithLogger()

var logDb = global.GvaEntLogDb

// getOpType 将 ent.Op 映射为 operationlog.OpType 的字符串值
func getOpType(op ent.Op) string {
	switch op {
	case ent.OpCreate:
		return operationlog.OpTypeDataCreate.String()
	case ent.OpUpdate, ent.OpUpdateOne:
		return operationlog.OpTypeDataUpdate.String()
	case ent.OpDelete, ent.OpDeleteOne:
		return operationlog.OpTypeDataDelete.String()
	default:
		return ""
	}
}

// getID: 适配 Int64 的通用 ID 获取
func getID[M ent.Mutation](ctx context.Context, op ent.Op, m M, result ent.Value) (int64, error) {
	// 1. 处理 Create 场景：从返回的 result 中获取 ID
	if op.Is(ent.OpCreate) && result != nil {
		// 先转为 any 再断言
		if v, ok := any(result).(interface{ GetID() int64 }); ok {
			return v.GetID(), nil
		}
	}

	// 2. 处理 Update/Delete 场景：从 Mutation 中获取 ID
	// 同样，必须先 any(m) 转换
	if v, ok := any(m).(interface{ ID() (int64, bool) }); ok {
		if val, exists := v.ID(); exists {
			return val, nil
		}
	}

	return 0, fmt.Errorf("无法获取对象 ID, Op: %s", op.String())
}

// getChangedFields: 返回 JSON 字符串以适配你的 Schema
func getChangedFields[M ent.Mutation](m M) (map[string]any, map[string]any) {
	changed := m.Fields()
	oldMap := make(map[string]any)
	newMap := make(map[string]any)

	// 使用软断言，防止断言失败导致 panic
	mutation, ok := any(m).(interface {
		OldField(context.Context, string) (any, error)
	})

	for _, f := range changed {
		// 只有断言成功才尝试获取旧值
		if ok {
			if val, err := mutation.OldField(context.Background(), f); err == nil {
				oldMap[f] = val
			}
		}
		// 获取新值
		if val, exists := m.Field(f); exists {
			newMap[f] = val
		}
	}

	return oldMap, newMap
}
