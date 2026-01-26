package types

import "vibe_ip/internal/ent_log"

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

	SystemVersion    string           `json:"system_version"`
	SystemGitCommit  string           `json:"system_git_commit"`
	SystemBuildTime  string           `json:"system_build_time"`
	SystemGoVersion  string           `json:"system_go_version"`
	SystemOS         string           `json:"system_os"`
	SystemArch       string           `json:"system_arch"`
	SystemMemoryInfo ReturnMemoryInfo `json:"memory_info"`

	Settings GetSettingResponse `json:"settings"`
}

// ReturnMemoryInfo 内存与 GC 综合信息
type ReturnMemoryInfo struct {
	AllocMB       float64 `json:"allocMB"`       // 已分配堆内存 (MB)
	SysMB         float64 `json:"sysMB"`         // 从系统获取的总内存 (MB)
	NumGC         uint32  `json:"numGC"`         // GC 总次数
	LastPauseMs   float64 `json:"lastPauseMs"`   // 最近一次 GC 暂停时间 (ms)
	PauseTotalMs  float64 `json:"pauseTotalMs"`  // GC 总暂停时间 (ms)
	GCCPUFraction float64 `json:"gccpuFraction"` // GC 占用 CPU 时间比例
	GCPerSec      float64 `json:"gcPerSec"`      // 平均每秒 GC 次数
	LastGCTime    string  `json:"lastGCTime"`    // 上次 GC 完成时间 (HH:MM:SS)
	HeapIdleMB    float64 `json:"heapIdleMB"`    // 空闲堆内存 (MB)
	StackInuseMB  float64 `json:"stackInuseMB"`  // 栈内存使用量 (MB)
	NumMallocs    uint64  `json:"numMallocs"`    // 总分配次数
	NumFrees      uint64  `json:"numFrees"`      // 总释放次数
	NextGCMB      float64 `json:"nextGCMB"`      // 下次GC触发的堆大小 (MB)
	HeapObjects   uint64  `json:"heapObjects"`   // 堆上活跃对象总数
	CPUCoreCount  int     `json:"cpuCoreCount"`  // CPU 核心数
	GoMaxProcs    int     `json:"goMaxProcs"`    // 当前 GOMAXPROCS 值
	NumGoroutine  int     `json:"numGoroutine"`  // 当前 Goroutine 数量

	RuntimeStartTime string `json:"runtimeStartTime"` // 应用启动时间
}
