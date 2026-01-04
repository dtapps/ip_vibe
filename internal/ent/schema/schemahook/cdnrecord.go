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

func CDNRecordLogHook() ent.Hook {
	return hook.On(
		func(next ent.Mutator) ent.Mutator {
			return hook.CDNRecordFunc(func(ctx context.Context, m *ent.CDNRecordMutation) (ent.Value, error) {
				start := time.Now()
				op := m.Op()
				opType := getOpType(op)
				objType := getObjType(m)

				// 1. Mutate 前：抓取变更字段和旧值快照
				oldV, _ := getChangedFields(ctx, m)

				// 2. 从 Context 获取来源标识
				source := GetSource(ctx)

				// 3. 执行数据库变更
				result, err := next.Mutate(ctx, m)

				// 4. 准备审计数据
				status := operationlog.StatusSuccess
				msg := "操作成功"
				if err != nil {
					status = operationlog.StatusFailed // 这里对应你 schematype 里的枚举值
					msg = err.Error()                  // 记录具体的报错原因
				}

				// 5. 获取 ID (如果是 Create 且报错了，ID 可能是 0)
				objID, _ := getObjID(ctx, op, m, result)

				// 6. 获取新值快照
				_, newV := getChangedFields(ctx, m)

				// 7. 异步记录审计日志
				go func(source string, id int64, ot string, oType string, ov map[string]any, nv map[string]any, stat string, message string, cost int64) {
					// 使用独立的 context 保证写入
					logErr := getLogDb().OperationLog.Create().
						SetSource(operationlog.Source(source)).
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
				}(source, objID, opType, objType, oldV, newV, string(status), msg, time.Since(start).Milliseconds())

				// 8. 本地打印
				if err != nil {
					getEntSlog().ErrorContext(ctx, fmt.Sprintf("%s 变更失败", m.Type()),
						slog.Int64("id", objID),
						slog.Any("err", err),
					)
				}

				// 依然返回原始的 result 和 err 给上层业务
				return result, err
			})
		},
		ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne|ent.OpDelete|ent.OpDeleteOne,
	)
}
