package schemahook

import (
	"context"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/esasiterecord"
	"vibe_ip/internal/ent/hook"
)

// EsaSiteRecordDataHook 为 EsaSiteRecord 添加数据处理钩子
func EsaSiteRecordDataHook() ent.Hook {
	return hook.On(
		func(next ent.Mutator) ent.Mutator {
			return hook.EsaSiteRecordFunc(func(ctx context.Context, m *ent.EsaSiteRecordMutation) (ent.Value, error) {

				// 替换 A/AAAA → A_AAAA
				if record_type, ok := m.RecordType(); ok {
					if record_type.String() == "A/AAAA" {
						m.SetRecordType(esasiterecord.RecordTypeA_AAAA)
					}
					if record_type == esasiterecord.RecordTypeA_AAAA {
						m.SetRecordType(esasiterecord.RecordTypeA_AAAA)
					}
				}

				// 继续执行后续操作
				return next.Mutate(ctx, m)
			})
		},
		ent.OpCreate|ent.OpUpdate|ent.OpUpdateOne,
	)
}
