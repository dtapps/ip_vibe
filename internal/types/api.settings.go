package types

// 获取配置 响应结构
type GetSettingResponse struct {
	DeviceOfflineInvalidTime int64  `json:"device_offline_invalid_time" jsonschema:"description=设备离线失效时间（分，默认 12 分钟，0表示不失效）"`   // 设备离线失效时间（分，默认 12 分钟，0表示不失效）
	IPReportIntervalTime     string `json:"ip_report_interval_time" jsonschema:"description=IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）"`    // IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
	PortForwardReloadTime    string `json:"port_forward_reload_time" jsonschema:"description=端口转发重载时间（分，默认 30 分钟，多个时用-分隔）"`     // 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
	PortForwardReportTime    string `json:"port_forward_report_time" jsonschema:"description=端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）"` // 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
}

// 更新设置 请求结构
type UpdateSettingRequest struct {
	GetSettingResponse
}
