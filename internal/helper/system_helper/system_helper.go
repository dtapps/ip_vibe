package system_helper

import (
	"context"
	"fmt"
	"runtime"
	"time"
	"vibe_ip/internal/global"
	"vibe_ip/internal/types"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemHelper struct{}

func NewSystemHelper() *SystemHelper {
	return &SystemHelper{}
}

// 记录应用启动时间
var runtimeStartTime = time.Now()

// GetMemoryInfo 获取系统信息和系统状态
func (s *SystemHelper) GetMemoryInfo(ctx context.Context) types.ReturnMemoryInfo {
	// 获取 Go 运行时内存 (虚拟)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 获取系统物理内存 (真实)
	v, _ := mem.VirtualMemoryWithContext(ctx)

	// 获取 CPU 芯片型号
	cpuInfo, _ := cpu.InfoWithContext(ctx)
	cpuModel := "Unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}

	return types.ReturnMemoryInfo{
		Version:   global.GvaVersion,
		GitCommit: global.GvaGitCommit,
		BuildTime: global.GvaBuildTime,
		GoVersion: global.GvaGoVersion,
		OS:        global.GvaOS,
		Arch:      global.GvaArch,

		RuntimeStartTime: runtimeStartTime.Format("01-02 15:04:05"),
		NumGoroutine:     runtime.NumGoroutine(),
		NumGC:            m.NumGC,

		CPUModel:    cpuModel,
		UsedPercent: v.UsedPercent,
		TotalMem:    fmt.Sprintf("%.1f GB", float64(v.Total)/1024/1024/1024),
		AllocMem:    fmt.Sprintf("%.1f MB", float64(m.Alloc)/1024/1024),
	}
}
