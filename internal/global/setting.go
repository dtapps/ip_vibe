package global

type Setting struct {
	// 设备离线失效时间（分，默认 12 分钟）
	DeviceOfflineInvalidTime int64 `json:"device_offline_invalid_time"`
	// IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
	IPReportIntervalTime string `json:"ip_report_interval_time"`
	// 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
	PortForwardReloadTime string `json:"port_reload_time"`
	// 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
	PortForwardReportTime string `json:"port_forward_report_time"`
	// CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	CDNRefreshTaskTime string `json:"cdn_refresh_task_time"`
	// DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	DNSRefreshTaskTime string `json:"dns_refresh_task_time"`
	// SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）
	SSLRefreshTaskTime string `json:"ssl_refresh_task_time"`
}
