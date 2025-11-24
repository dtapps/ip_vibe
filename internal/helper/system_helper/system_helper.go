package system_helper

import (
	"runtime"
	"time"
)

type SystemHelper struct{}

func NewSystemHelper() *SystemHelper {
	return &SystemHelper{}
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

// 记录应用启动时间（用于计算 gcPerSec）
var runtimeStartTime = time.Now()

// GetMemoryInfo 返回内存和 GC 综合信息
func (s *SystemHelper) GetMemoryInfo() ReturnMemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算最近一次 GC 暂停时间（毫秒）
	lastPauseMs := float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6

	// 计算平均每秒 GC 次数（基于运行时间）
	uptimeSec := time.Since(runtimeStartTime).Seconds()
	gcPerSec := 0.0
	if uptimeSec > 0 {
		gcPerSec = float64(m.NumGC) / uptimeSec
	}

	// 返回一个完整的 ReturnMemoryInfo 结构体
	return ReturnMemoryInfo{
		AllocMB:       float64(m.Alloc) / 1024 / 1024,
		SysMB:         float64(m.Sys) / 1024 / 1024,
		NumGC:         m.NumGC,
		LastPauseMs:   lastPauseMs,
		PauseTotalMs:  float64(m.PauseTotalNs) / 1e6,
		GCCPUFraction: m.GCCPUFraction,
		GCPerSec:      gcPerSec,
		LastGCTime:    time.Unix(0, int64(m.LastGC)).Format("15:04:05"),
		HeapIdleMB:    float64(m.HeapIdle) / 1024 / 1024,
		StackInuseMB:  float64(m.StackInuse) / 1024 / 1024,
		NumMallocs:    m.Mallocs,
		NumFrees:      m.Frees,
		NextGCMB:      float64(m.NextGC) / 1024 / 1024,
		HeapObjects:   m.HeapObjects,
		CPUCoreCount:  runtime.NumCPU(),
		GoMaxProcs:    runtime.GOMAXPROCS(0),
		NumGoroutine:  runtime.NumGoroutine(),

		RuntimeStartTime: runtimeStartTime.Format("01-02 15:04:05"),
	}
}
