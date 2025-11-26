package operation_logs_helper

import (
	"context"
	"log/slog"
	"vibe_ip/ent"
	"vibe_ip/ent/operationlog"
)

type OperationLogHelper struct {
	db        *ent.Client
	dbEnabled bool
	tx        *ent.Tx
	txEnabled bool
}

func NewOperationLogHelper() *OperationLogHelper {
	return &OperationLogHelper{}
}

// SetDb 设置数据库客户端
func (s *OperationLogHelper) SetDb(db *ent.Client) *OperationLogHelper {
	s.db = db
	s.dbEnabled = true
	return s
}

// SetTx 设置事务
func (s *OperationLogHelper) SetTx(tx *ent.Tx) *OperationLogHelper {
	s.tx = tx
	s.txEnabled = true
	return s
}

// RecordDataCreate 记录数据创建操作日志
func (s *OperationLogHelper) RecordDataCreate(ctx context.Context, obj_id int64, obj_type operationlog.ObjType, status operationlog.Status, message string) error {
	var create *ent.OperationLogCreate
	if s.db == nil && s.dbEnabled {
		create = s.db.OperationLog.Create()
	} else if s.tx != nil && s.txEnabled {
		create = s.tx.OperationLog.Create()
	} else {
		slog.ErrorContext(ctx, "记录数据创建操作日志失败",
			slog.String("err", "数据库客户端和事务都未启用"),
		)
		return nil
	}

	err := create.SetOpType(operationlog.OpTypeDataCreate).
		SetObjID(obj_id).
		SetObjType(obj_type).
		SetStatus(status).
		SetMessage(message).
		Exec(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "记录数据创建操作日志失败",
			slog.Any("err", err),
		)
	}
	return err
}

// RecordDataUpdate 记录数据更新操作日志
func (s *OperationLogHelper) RecordDataUpdate(ctx context.Context, obj_id int64, obj_type operationlog.ObjType, status operationlog.Status, message string) error {
	var create *ent.OperationLogCreate
	if s.db == nil && s.dbEnabled {
		create = s.db.OperationLog.Create()
	} else if s.tx != nil && s.txEnabled {
		create = s.tx.OperationLog.Create()
	} else {
		slog.ErrorContext(ctx, "记录数据更新操作日志失败",
			slog.String("err", "数据库客户端和事务都未启用"),
		)
		return nil
	}

	err := create.SetOpType(operationlog.OpTypeDataUpdate).
		SetObjID(obj_id).
		SetObjType(obj_type).
		SetStatus(status).
		SetMessage(message).
		Exec(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "记录数据更新操作日志失败",
			slog.Any("err", err),
		)
	}
	return err
}

// RecordDataDelete 记录数据删除操作日志
func (s *OperationLogHelper) RecordDataDelete(ctx context.Context, obj_id int64, obj_type operationlog.ObjType, status operationlog.Status, message string) error {
	var create *ent.OperationLogCreate
	if s.db == nil && s.dbEnabled {
		create = s.db.OperationLog.Create()
	} else if s.tx != nil && s.txEnabled {
		create = s.tx.OperationLog.Create()
	} else {
		slog.ErrorContext(ctx, "记录数据删除操作日志失败",
			slog.String("err", "数据库客户端和事务都未启用"),
		)
		return nil
	}

	err := create.SetOpType(operationlog.OpTypeDataDelete).
		SetObjID(obj_id).
		SetObjType(obj_type).
		SetStatus(status).
		SetMessage(message).
		Exec(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "记录数据删除操作日志失败",
			slog.Any("err", err),
		)
	}
	return err
}
