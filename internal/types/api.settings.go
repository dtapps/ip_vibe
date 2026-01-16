package types

// 获取设置 响应结构
type GetSettingResponse struct {
	DeviceOfflineInvalidTime int64  `json:"device_offline_invalid_time" jsonschema:"description=设备离线失效时间（分，默认 12 分钟）"`          // 设备离线失效时间（分，默认 12 分钟）
	IPReportIntervalTime     string `json:"ip_report_interval_time" jsonschema:"description=IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）"`    // IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
	PortForwardReloadTime    string `json:"port_forward_reload_time" jsonschema:"description=端口转发重载时间（分，默认 30 分钟，多个时用-分隔）"`     // 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
	PortForwardReportTime    string `json:"port_forward_report_time" jsonschema:"description=端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）"` // 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
	CDNRefreshTaskTime       string `json:"cdn_refresh_task_time" jsonschema:"description=CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）"`    // CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	DNSRefreshTaskTime       string `json:"dns_refresh_task_time" jsonschema:"description=DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）"`    // DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	SSLRefreshTaskTime       string `json:"ssl_refresh_task_time" jsonschema:"description=SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）"`     // SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）
}

// 更新设置 请求结构
type UpdateSettingRequest struct {
	DeviceOfflineInvalidTime int64  `json:"device_offline_invalid_time,omitempty" jsonschema:"description=设备离线失效时间（分，默认 12 分钟）"`                // 设备离线失效时间（分，默认 12 分钟）
	IPReportIntervalTime     string `json:"ip_report_interval_time" binding:"required" jsonschema:"description=IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）"` // IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
	PortForwardReloadTime    string `json:"port_forward_reload_time,omitempty" jsonschema:"description=端口转发重载时间（分，默认 30 分钟，多个时用-分隔）"`           // 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
	PortForwardReportTime    string `json:"port_forward_report_time,omitempty" jsonschema:"description=端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）"`       // 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
	CDNRefreshTaskTime       string `json:"cdn_refresh_task_time" binding:"required" jsonschema:"description=CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）"` // CDN刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	DNSRefreshTaskTime       string `json:"dns_refresh_task_time" binding:"required" jsonschema:"description=DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）"` // DNS刷新任务时间（分，默认 10-12 分钟，多个时用-分隔）
	SSLRefreshTaskTime       string `json:"ssl_refresh_task_time" binding:"required" jsonschema:"description=SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）"`  // SSL刷新任务时间（时，默认 8-10 小时，多个时用-分隔）
}
