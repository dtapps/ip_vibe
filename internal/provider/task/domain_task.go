package task

// DomainMonitorTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type DomainMonitorTaskArgs struct {
	ID int64 `json:"id"` // DomainMonitor 的数据库 ID
}
