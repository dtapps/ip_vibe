package task

// IPReportTaskConfig 配置
type IPReportTaskConfig struct {
	DeviceCode string `json:"device_code"` // 设备编号
	DeviceName string `json:"device_name"` // 设备名称

	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}
