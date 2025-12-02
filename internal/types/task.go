package types

import (
	"vibe_ip/internal/helper/gocron_helper"
)

// 任务页面 请求结构
type TaskRequest struct {
	Page     int `form:"page" default:"1"`       // 当前页码
	PageSize int `form:"page_size" default:"10"` // 每页数量
}

// 任务页面 响应结构
type TaskResponse struct {
	List     []gocron_helper.JobData `json:"list"`      // 任务列表
	Page     int                     `json:"page"`      // 当前页码
	PageSize int                     `json:"page_size"` // 每页数量
	Total    int64                   `json:"total"`     // 总数量
	Status   string                  `json:"status"`    // 任务运行状态
}
