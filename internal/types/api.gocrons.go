package types

// 获取全部任务 请求结构
type GetGocronsRequest struct {
}

// 获取任务 请求结构
type GetGocronRequest struct {
	ID string `uri:"id" json:"id" binding:"required" jsonschema:"任务的唯一标识符"` // 任务序号
}

// 运行任务 请求结构
type RunGocronRequest struct {
	GetGocronRequest
}

// 删除任务 请求结构
type DeleteGocronRequest struct {
	GetGocronRequest
}
