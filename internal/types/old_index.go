package types

import (
	"vibe_ip/internal/helper/gocron_helper"
	"vibe_ip/internal/helper/system_helper"
)

type OldIndexResponse struct {
	Configs      []GetConfigResponse      `json:"configs"`
	CdnRecords   []GetCdnRecordResponse   `json:"cdn_records"`
	DnsRecords   []GetDnsRecordResponse   `json:"dns_records"`
	Devices      []GetDeviceResponse      `json:"devices"`
	IpGroups     []GetIPGroupResponse     `json:"ip_groups"`
	PortForwards []GetPortForwardResponse `json:"port_forwards"`
	Webhooks     []GetWebhookResponse     `json:"webhooks"`
	WebhookSubs  []GetWebhookSubResponse  `json:"webhook_subscriptions"`
	Tasks        []gocron_helper.JobData  `json:"tasks"`

	Logs     []GetLogResponse `json:"logs"`
	LogLimit int              `json:"log_limit"`

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
