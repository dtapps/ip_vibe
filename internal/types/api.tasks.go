package types

// 获取任务 请求结构
type GetTaskRequest struct {
	ID string `uri:"id" json:"-" binding:"required" jsonschema:"title=任务序号,description=任务的唯一标识符,example=task-001"` // 任务序号
}

// 运行任务 请求结构
type RunTaskRequest struct {
	GetTaskRequest
}

// 删除任务 请求结构
type DeleteTaskRequest struct {
	GetTaskRequest
}
