package task

// SSLApplTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type SSLApplTaskArgs struct {
	ID int64 `json:"id"` // SSLAppl 的数据库 ID
}
