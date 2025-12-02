package types

import (
	"vibe_ip/internal/helper/gocron_helper"
)

type TaskResponse struct {
	List     []gocron_helper.JobData `json:"list"`      // 任务列表
	Page     int                     `json:"page"`      // 当前页码
	PageSize int                     `json:"page_size"` // 每页数量
	Total    int64                   `json:"total"`     // 总数量
	Status   string                  `json:"status"`    // 任务运行状态
}
