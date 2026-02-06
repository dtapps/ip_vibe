package types

import "time"

// 任务页面 请求结构
type GocronRequest struct {
	CommonPageRequest
}

// 任务页面 响应结构
type GocronResponse struct {
	CommonPageResponse[GocronJob, GocronRequest]        // 任务列表
	Status                                       string `json:"status"` // 任务运行状态
}

type GocronJobInfo struct {
	ID      string    `json:"id"`       // 任务序号
	Name    string    `json:"name"`     // 任务名称
	LastRun time.Time `json:"last_run"` // 上次运行时间
	NextRun time.Time `json:"next_run"` // 下次运行时间
}

type GocronJob struct {
	ID             string      `json:"id"`              // 任务序号
	Name           string      `json:"name"`            // 任务名称
	LastRun        time.Time   `json:"last_run"`        // 上次运行时间
	NextRun        time.Time   `json:"next_run"`        // 下次运行时间
	ExecutionTime  []time.Time `json:"execution_time"`  // 执行时间列表
	Schedule       string      `json:"schedule"`        // 任务调度表达式
	ScheduleDetail string      `json:"schedule_detail"` // 任务调度表达式详情
}
