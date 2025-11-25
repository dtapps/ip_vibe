package types

import (
	"time"
)

// 获取全部操作日志 响应结构
type GetOperationLogsResponse struct {
	ID        int64     `json:"id"`         // IP组序号
	OpType    string    `json:"op_type"`    // 操作类型
	ObjID     int64     `json:"obj_id"`     // 对象主键
	ObjType   string    `json:"obj_type"`   // 对象类型
	OldValue  string    `json:"old_value"`  // 旧值
	NewValue  string    `json:"new_value"`  // 新值
	Status    string    `json:"status"`     // 状态
	Message   string    `json:"message"`    // 详情
	CreatedAt time.Time `json:"created_at"` // 创建时间
}
