package types

// 日志页面 请求结构
type LogRequest struct {
	CommonPageRequest
	ID      int64  `json:"id,omitempty" form:"id,omitempty"`             // 日志序号
	OpType  string `json:"op_type,omitempty" form:"op_type,omitempty"`   // 操作类型
	ObjType string `json:"obj_type,omitempty" form:"obj_type,omitempty"` // 对象类型
	ObjID   int64  `json:"obj_id,omitempty" form:"obj_id,omitempty"`     // 对象主键
	Status  string `json:"status,omitempty" form:"status,omitempty"`     // 状态
	Message string `json:"message,omitempty" form:"message,omitempty"`   // 详情
}

// 日志页面 响应结构
type LogResponse struct {
	CommonPageResponse[GetLogResponse, LogRequest] // 操作日志列表
}
