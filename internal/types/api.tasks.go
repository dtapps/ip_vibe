package types

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// 获取全部任务 请求结构
type GetTasksRequest struct {
	// 任务状态
	Status string `json:"status,omitempty" jsonschema:"任务状态"`
}

// 获取任务详情 响应结构
type GetTaskResponse struct {
	// 任务序号
	ID int64 `json:"id"`
	// 任务编号
	EntryID uuid.UUID `json:"entry_id"`
	// 任务名称
	Name string `json:"name"`
	// 任务处理函数
	HandlerName string `json:"handler_name"`
	// 任务处理函数参数
	HandlerArgs json.RawMessage `json:"handler_args"`
	// 任务类型: one_off 一次性, recurring 周期性
	TaskType string `json:"task_type"`
	// 任务类型: cron cron 表达式, duration 固定间隔, random 随机间隔
	JobType string `json:"job_type"`
	// 任务分类: system 系统内置, custom 用户自定义
	JobCategory string `json:"job_category"`
	// 业务类型: cdn, dns, ssl, system
	BusinessType string `json:"business_type"`
	// 业务序号
	BusinessID int64 `json:"business_id"`
	// cron 表达式
	CronExpr string `json:"cron_expr"`
	// 是否包含秒
	WithSeconds bool `json:"with_seconds"`
	// 固定间隔值 (配合 interval_unit 使用)
	IntervalValue int `json:"interval_value"`
	// 随机最小值 (配合 interval_unit 使用)
	MinInterval int `json:"min_interval"`
	// 随机最大值 (配合 interval_unit 使用)
	MaxInterval int `json:"max_interval"`
	// 间隔时间单位 (用于 duration 和 random 模式)
	IntervalUnit string `json:"interval_unit"`
	// 状态: running 运行中, stopped 已停止
	Status string `json:"status"`
	// 是否单例
	IsSingleton bool `json:"is_singleton"`

	CreatedAt time.Time `json:"-"` // 创建时间
	UpdatedAt time.Time `json:"-"` // 更新时间
}

// 获取任务 请求结构
type GetTaskRequest struct {
	ID int64 `uri:"id" json:"id" binding:"required" jsonschema:"任务序号"` // 任务序号
}

// 更新任务 请求结构
type UpdateTaskRequest struct {
	GetTaskRequest
	JobType       string `json:"job_type" binding:"required" jsonschema:"任务类型"`    // 任务类型
	CronExpr      string `json:"cron_expr,omitempty" jsonschema:"cron 表达式"`        // cron 表达式
	WithSeconds   bool   `json:"with_seconds,omitempty" jsonschema:"是否包含秒"`        // 是否包含秒
	IntervalValue int    `json:"interval_value,omitempty" jsonschema:"固定间隔值"`      // 固定间隔值
	MinInterval   int    `json:"min_interval,omitempty" jsonschema:"随机最小值"`        // 随机最小值
	MaxInterval   int    `json:"max_interval,omitempty" jsonschema:"随机最大值"`        // 随机最大值
	IntervalUnit  string `json:"interval_unit,omitempty" jsonschema:"间隔时间单位"`      // 间隔时间单位
	IsSingleton   bool   `json:"is_singleton,omitempty" jsonschema:"是否单例"`         // 是否单例
	RunOnStart    bool   `json:"run_on_start,omitempty" jsonschema:"是否在启动时立即执行一次"` // 是否在启动时立即执行一次
}

// 更新任务状态 请求结构
type UpdateTaskActiveRequest struct {
	GetTaskRequest
	Status string `json:"status,omitempty" jsonschema:"任务状态"` // 任务状态
}

// 批量更新任务状态 请求结构
type BatchUpdateTaskActiveRequest struct {
	BatchDeleteTaskRequest
	Status string `json:"status,omitempty" jsonschema:"任务状态"` // 任务状态
}

// 删除任务 请求结构
type DeleteTaskRequest struct {
	GetTaskRequest
}

// 批量删除任务 请求结构
type BatchDeleteTaskRequest struct {
	IDS []int64 `json:"ids" binding:"required" jsonschema:"任务序号列表"` // 任务序号列表
}
