package types

import (
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/system_helper"
)

type Indexesponse struct {
	Configs           []GetConfigsResponse       `json:"configs"`
	CdnRecords        []GetCdnRecordsResponse    `json:"cdn_records"`
	DnsRecords        []GetDnsRecordsResponse    `json:"dns_records"`
	Devices           []GetDevicesResponse       `json:"devices"`
	IpGroups          []GetIPGroupsResponse      `json:"ip_groups"`
	Tasks             []gocron_helper.JobData    `json:"tasks"`
	OperationLogs     []GetOperationLogsResponse `json:"operation_logs"`
	OperationLogLimit int                        `json:"operation_log_limit"`

	TaskStatus string `json:"task_status"`
	DeviceCode string `json:"device_code"`

	SystemVersion    string                         `json:"system_version"`
	SystemGitCommit  string                         `json:"system_git_commit"`
	SystemBuildTime  string                         `json:"system_build_time"`
	SystemGoVersion  string                         `json:"system_go_version"`
	SystemOS         string                         `json:"system_os"`
	SystemArch       string                         `json:"system_arch"`
	SystemMemoryInfo system_helper.ReturnMemoryInfo `json:"memory_info"`
}
