package types

// 任务页面 请求结构
type TaskRequest struct {
	CommonPageRequest
	ID           int64  `json:"id,omitempty" form:"id,omitempty"`                       // 任务序号
	EntryID      string `json:"entry_id,omitempty" form:"entry_id,omitempty"`           // 任务编号
	Name         string `json:"name,omitempty" form:"name,omitempty"`                   // 任务名称
	TaskType     string `json:"task_type,omitempty" form:"task_type,omitempty"`         // 任务类型
	JobType      string `json:"job_type,omitempty" form:"job_type,omitempty"`           // 任务类型
	JobCategory  string `json:"job_category,omitempty" form:"job_category,omitempty"`   // 任务分类
	BusinessType string `json:"business_type,omitempty" form:"business_type,omitempty"` // 业务类型
	BusinessID   int64  `json:"business_id,omitempty" form:"business_id,omitempty"`     // 业务序号
	Status       string `json:"status,omitempty" form:"status,omitempty"`               // 状态
}

// 任务页面 响应结构
type TaskResponse struct {
	CommonPageResponse[GetTaskResponse, TaskRequest] // 任务列表
}
