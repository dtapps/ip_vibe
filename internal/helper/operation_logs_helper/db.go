package operation_logs_helper

import (
	"context"
	"log/slog"
	"vibe_ip/ent"
	"vibe_ip/ent/operationlog"
)

type OperationLogHelper struct {
	db *ent.Client
	tx *ent.Tx
}

func NewOperationLogHelper(db *ent.Client) *OperationLogHelper {
	return &OperationLogHelper{db: db}
}

// SetTx 设置事务
func (s *OperationLogHelper) SetTx(tx *ent.Tx) *OperationLogHelper {
	s.tx = tx
	return s
}

// GetAll 获取所有操作日志
func (s *OperationLogHelper) GetAll(ctx context.Context) ([]*ent.OperationLog, error) {
	query := s.db.OperationLog.Query()
	if s.tx != nil {
		query = s.tx.OperationLog.Query()
	}
	return query.All(ctx)
}

// RecordDataCreate 记录数据创建操作日志
func (s *OperationLogHelper) RecordDataCreate(ctx context.Context, obj_id int64, obj_type operationlog.ObjType, status operationlog.Status, message string) error {
	create := s.db.OperationLog.Create()
	if s.tx != nil {
		create = s.tx.OperationLog.Create()
	}
	err := create.
		SetOpType(operationlog.OpTypeDataCreate).
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
	create := s.db.OperationLog.Create()
	if s.tx != nil {
		create = s.tx.OperationLog.Create()
	}
	err := create.
		SetOpType(operationlog.OpTypeDataUpdate).
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
	create := s.db.OperationLog.Create()
	if s.tx != nil {
		create = s.tx.OperationLog.Create()
	}
	err := create.
		SetOpType(operationlog.OpTypeDataDelete).
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
