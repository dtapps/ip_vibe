package schemahook

import (
	"context"
	"fmt"
	"log/slog"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent_log"
	"vibe_ip/internal/ent_log/operationlog"
	"vibe_ip/internal/global"
)

func getSlog() *slog.Logger {
	entSlog := slog.Default()
	if global.GvaEntSlog != nil {
		entSlog = global.GvaEntSlog.WithLogger()
	}
	return entSlog
}

func getLogDb() (logDb *ent_log.Client) {
	if logDb == nil {
		logDb = global.GvaEntLogDb
	}
	return logDb
}

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

// 获取表名逻辑
func getObjType[M ent.Mutation](m M) string {
	var entityToObjType = map[string]operationlog.ObjType{
		ent.TypeCDNRecord:           operationlog.ObjTypeCdnRecord,
		ent.TypeDNSRecord:           operationlog.ObjTypeDNSRecord,
		ent.TypeDevice:              operationlog.ObjTypeDevice,
		ent.TypeDevicePortForward:   operationlog.ObjTypeDevicePortForward,
		ent.TypeDomainMonitor:       operationlog.ObjTypeDomainMonitor,
		ent.TypeIpGroup:             operationlog.ObjTypeIPGroup,
		ent.TypeIpGroupDevice:       operationlog.ObjTypeIPGroupDevice,
		ent.TypePlatformConfig:      operationlog.ObjTypePlatformConfig,
		ent.TypeSSLRecord:           operationlog.ObjTypeSslRecord,
		ent.TypeWebhook:             operationlog.ObjTypeWebhook,
		ent.TypeWebhookSubscription: operationlog.ObjTypeWebhookSubscription,

		ent_log.TypeDeviceIpReport:        operationlog.ObjTypeDeviceIPReport,
		ent_log.TypeDomainMonitorHistory:  operationlog.ObjTypeDomainMonitorHistory,
		ent_log.TypeHttpLog:               operationlog.ObjTypeHTTPLog,
		ent_log.TypeOperationLog:          operationlog.ObjTypeOperationLog,
		ent_log.TypePortForwardStats:      operationlog.ObjTypePortForwardStats,
		ent_log.TypeRestyLog:              operationlog.ObjTypeRestyLog,
		ent_log.TypeSSLCertificateHistory: operationlog.ObjTypeSslCertificateHistory,
		ent_log.TypeWebhookHistory:        operationlog.ObjTypeWebhookHistory,
	}
	typeName := m.Type()
	if val, ok := entityToObjType[typeName]; ok {
		return val.String()
	}
	return typeName
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
