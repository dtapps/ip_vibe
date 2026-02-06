package task

// DNSTaskArgs 定义了存放在 CronJob.handler_args 中的数据结构
type DNSTaskArgs struct {
	ID int64 `json:"id"` // DNSRecord 的数据库 ID
}
