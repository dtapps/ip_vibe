package types

import (
	"vibe_ip/internal/helper/gocron_helper"
)

// 任务页面 请求结构
type TaskRequest struct {
	CommonPageRequest
}

// 任务页面 响应结构
type TaskResponse struct {
	CommonPageResponse[gocron_helper.JobData]        // 任务列表
	Status                                    string `json:"status"` // 任务运行状态
}
