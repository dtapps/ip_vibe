package types

import "vibe_ip/internal/helper/gocron_helper"

type Indexesponse struct {
	Configs       []GetConfigsResponse       `json:"configs"`
	CdnRecords    []GetCdnRecordsResponse    `json:"cdn_records"`
	DnsRecords    []GetDnsRecordsResponse    `json:"dns_records"`
	Devices       []GetDevicesResponse       `json:"devices"`
	IpGroups      []GetIPGroupsResponse      `json:"ip_groups"`
	Tasks         []gocron_helper.JobData    `json:"tasks"`
	OperationLogs []GetOperationLogsResponse `json:"operation_logs"`

	TaskStatus string `json:"task_status"`
	DeviceCode string `json:"device_code"`
}
