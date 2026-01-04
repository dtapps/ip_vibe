package schemahook

import (
	"context"
	"fmt"
	"log/slog"
	"time"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/hook"
	"vibe_ip/internal/ent_log/operationlog"
)

func WebhookLogHook() ent.Hook {
	return hook.On(
		func(next ent.Mutator) ent.Mutator {
			return hook.WebhookFunc(func(ctx context.Context, m *ent.WebhookMutation) (ent.Value, error) {
				start := time.Now()
				op := m.Op()
				opType := getOpType(op)
				objType := getObjType(m)

				// 1. Mutate 前：抓取变更字段和旧值快照
				oldV, _ := getChangedFields(ctx, m)

				// 2. 执行数据库变更
				result, err := next.Mutate(ctx, m)

				// 3. 准备审计数据
				status := operationlog.StatusSuccess
				msg := "操作成功"
				if err != nil {
					status = operationlog.StatusFailed // 这里对应你 schematype 里的枚举值
					msg = err.Error()                  // 记录具体的报错原因
				}

				// 4. 获取 ID (如果是 Create 且报错了，ID 可能是 0)
				objID, _ := getObjID(ctx, op, m, result)

				// 5. 获取新值快照
				_, newV := getChangedFields(ctx, m)

				// 6. 异步记录审计日志
				go func(id int64, ot string, oType string, ov map[string]any, nv map[string]any, stat string, message string, cost int64) {
					// 使用独立的 context 保证写入
					logErr := getLogDb().OperationLog.Create().
						SetOpType(operationlog.OpType(ot)).
						SetObjType(operationlog.ObjType(oType)).
						SetObjID(id).
						SetOldValue(ov).
						SetNewValue(nv).
						SetStatus(operationlog.Status(stat)).
						SetMessage(message).
						Exec(context.Background())
					if logErr != nil {
						getEntLogSlog().ErrorContext(ctx, fmt.Sprintf("%s 记录操作日志失败", m.Type()),
							slog.Any("err", logErr),
						)
					}
				}(objID, opType, objType, oldV, newV, string(status), msg, time.Since(start).Milliseconds())

				// 7. 本地打印
				if err != nil {
					getEntSlog().ErrorContext(ctx, fmt.Sprintf("%s 变更失败", m.Type()),
						slog.Int64("id", objID),
						slog.Any("err", err),
					)
				}

				return result, err // 依然返回原始的 result 和 err 给上层业务
			})
		},
		ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
	)
}
