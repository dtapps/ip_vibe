package types

import (
	"time"
)

// 获取操作日志 响应结构
type GetLogResponse struct {
	ID        int64     `json:"id" jsonschema:"title=日志序号,description=日志记录的唯一标识符"`              // 日志序号
	OpType    string    `json:"op_type" jsonschema:"title=操作类型,description=执行的操作类型"`             // 操作类型
	ObjID     int64     `json:"obj_id" jsonschema:"title=对象主键,description=操作对象的主键ID"`            // 对象主键
	ObjType   string    `json:"obj_type" jsonschema:"title=对象类型,description=操作对象的类型"`            // 对象类型
	OldValue  string    `json:"old_value" jsonschema:"title=旧值,description=操作前的值"`                // 旧值
	NewValue  string    `json:"new_value" jsonschema:"title=新值,description=操作后的值"`                // 新值
	Status    string    `json:"status" jsonschema:"title=状态,description=操作执行的状态,enum=success,enum=failed"` // 状态
	Message   string    `json:"message" jsonschema:"title=详情,description=操作的详细信息"`               // 详情
	CreatedAt time.Time `json:"created_at" jsonschema:"title=创建时间,description=日志记录的创建时间"`        // 创建时间
}
