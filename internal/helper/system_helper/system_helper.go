package system_helper

import (
	"runtime"
	"time"
	"vibe_ip/internal/types"
)

type SystemHelper struct{}

func NewSystemHelper() *SystemHelper {
	return &SystemHelper{}
}

// 记录应用启动时间（用于计算 gcPerSec）
var runtimeStartTime = time.Now()

// GetMemoryInfo 返回内存和 GC 综合信息
func (s *SystemHelper) GetMemoryInfo() types.ReturnMemoryInfo {
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
	return types.ReturnMemoryInfo{
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
