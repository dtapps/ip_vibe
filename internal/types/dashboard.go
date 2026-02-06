package types

import (
	"time"
	"vibe_ip/internal/ent_log"
)

// 仪表盘 统计
type DashboardStat struct {
	Label string `json:"label"` // 标题
	Value int    `json:"value"` // 数值
}

// 仪表盘 快速入口
type DashboardEntrance struct {
	Title string `json:"title"` // 标题
	Icon  string `json:"icon"`  // 图标
	Href  string `json:"href"`  // 跳转链接
}

// 仪表盘 响应结构
type DashboardResponse struct {
	Stats      []DashboardStat        `json:"stats"`
	Entrances  []DashboardEntrance    `json:"entrances"`
	RecentLogs []ent_log.OperationLog `json:"recent_logs"`

	DeviceCode string `json:"device_code"`

	SystemInfo       SystemInfo       `json:"system_info"`
	SystemMemoryInfo ReturnMemoryInfo `json:"memory_info"`

	Settings GetSettingResponse `json:"settings"`
}

type SystemInfo struct {
	Version          string    `json:"version"`            // 系统版本
	GitCommit        string    `json:"git_commit"`         // 系统 Git 提交
	BuildTime        string    `json:"build_time"`         // 系统构建时间
	GoVersion        string    `json:"go_version"`         // 系统 Go 版本
	OS               string    `json:"os"`                 // 系统 OS
	Arch             string    `json:"arch"`               // 系统架构
	RuntimeStartTime time.Time `json:"runtime_start_time"` // 应用启动时间
}

type ReturnMemoryInfo struct {
	NumGoroutine int    `json:"num_goroutine"` // Goroutine 数量
	NumGC        uint32 `json:"num_gc"`        // GC 总次数

	CPUModel    string  `json:"cpu_model"`    // CPU 型号
	UsedPercent float64 `json:"used_percent"` // 内存使用率
	TotalMem    string  `json:"total_mem"`    // 总物理内存
	AllocMem    string  `json:"alloc_mem"`    // Go 进程分配内存
}
