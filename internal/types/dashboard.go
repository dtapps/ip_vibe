package types

import "vibe_ip/internal/helper/system_helper"

// 仪表盘 响应结构
type DashboardResponse struct {
	Stats struct {
		DeviceCount       int
		OnlineDeviceCount int
		DnsRecordCount    int
		CdnRecordCount    int
		WebhookCount      int
		ActiveTaskCount   int
		LogLimit          int
	}
	RecentLogs []GetLogResponse

	DeviceCode string `json:"device_code"`

	SystemVersion    string                         `json:"system_version"`
	SystemGitCommit  string                         `json:"system_git_commit"`
	SystemBuildTime  string                         `json:"system_build_time"`
	SystemGoVersion  string                         `json:"system_go_version"`
	SystemOS         string                         `json:"system_os"`
	SystemArch       string                         `json:"system_arch"`
	SystemMemoryInfo system_helper.ReturnMemoryInfo `json:"memory_info"`
}
