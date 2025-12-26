package types

// 任务页面 请求结构
type TaskRequest struct {
	CommonPageRequest
}

// 任务页面 响应结构
type TaskResponse struct {
	CommonPageResponse[JobData]        // 任务列表
	Status                      string `json:"status"` // 任务运行状态
}

type JobData struct {
	ID             string   `json:"id"`              // 任务序号
	Name           string   `json:"name"`            // 任务名称
	LastRun        string   `json:"last_run"`        // 上次运行时间
	NextRun        string   `json:"next_run"`        // 下次运行时间
	NextRuns       []string `json:"next_runs"`       // 下次运行时间列表
	Schedule       string   `json:"schedule"`        // 任务调度表达式
	ScheduleDetail string   `json:"schedule_detail"` // 任务调度表达式详情
}

type JobInfo struct {
	ID      string `json:"id"`       // 任务序号
	Name    string `json:"name"`     // 任务名称
	LastRun string `json:"last_run"` // 上次运行时间
	NextRun string `json:"next_run"` // 下次运行时间
}
