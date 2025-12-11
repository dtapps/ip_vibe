package types

import (
	"time"
)

// 获取操作日志 响应结构
type GetLogResponse struct {
	ID        int64     `json:"id" jsonschema:"日志记录的唯一标识符"`        // 日志序号
	OpType    string    `json:"op_type" jsonschema:"执行的操作类型"`      // 操作类型
	ObjID     int64     `json:"obj_id" jsonschema:"操作对象的主键ID"`     // 对象主键
	ObjType   string    `json:"obj_type" jsonschema:"操作对象的类型"`     // 对象类型
	OldValue  string    `json:"old_value" jsonschema:"操作前的值"`      // 旧值
	NewValue  string    `json:"new_value" jsonschema:"操作后的值"`      // 新值
	Status    string    `json:"status" jsonschema:"操作执行的状态"`       // 状态
	Message   string    `json:"message" jsonschema:"操作的详细信息"`      // 详情
	CreatedAt time.Time `json:"created_at" jsonschema:"日志记录的创建时间"` // 创建时间
}
