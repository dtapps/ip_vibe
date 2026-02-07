package types

import (
	"vibe_ip/internal/ent_log"
)

// 任务日志页面 请求结构
type TaskLogRequest struct {
	CommonPageRequest
	ID        int64  `json:"id,omitempty" form:"id,omitempty"`                   // 任务序号
	CronJobID int64  `json:"cron_job_id,omitempty" form:"cron_job_id,omitempty"` // 关联 定时任务 序号
	Status    string `json:"status,omitempty" form:"status,omitempty"`           // 状态
	EntryID   string `json:"entry_id,omitempty" form:"entry_id,omitempty"`       // 任务编号
}

// 任务日志页面 响应结构
type TaskLogResponse struct {
	CommonPageResponse[ent_log.CronJobLog, TaskLogRequest] // 任务列表
}
