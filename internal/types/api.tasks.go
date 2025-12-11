package types

// 获取全部任务 请求结构
type GetTasksRequest struct {
}

// 获取任务 请求结构
type GetTaskRequest struct {
	ID string `uri:"id" json:"id" binding:"required" jsonschema:"任务的唯一标识符"` // 任务序号
}

// 运行任务 请求结构
type RunTaskRequest struct {
	GetTaskRequest
}

// 删除任务 请求结构
type DeleteTaskRequest struct {
	GetTaskRequest
}
