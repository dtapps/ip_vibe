package types

// 获取任务 请求结构
type GetTaskRequest struct {
	ID string `uri:"id" binding:"required"` // 任务序号
}

// 运行任务 请求结构
type RunTaskRequest struct {
	GetTaskRequest
}

// 删除任务 请求结构
type DeleteTaskRequest struct {
	GetTaskRequest
}
