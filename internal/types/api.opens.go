package types

// 获取配置 响应结构
type OpenGetSettingResponse struct {
	IPReportIntervalTime  string `json:"ip_report_interval_time" jsonschema:"description=IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）"`    // IP上报间隔时间（分，默认 8-12 分钟，多个时用-分隔）
	PortForwardReloadTime string `json:"port_forward_reload_time" jsonschema:"description=端口转发重载时间（分，默认 30 分钟，多个时用-分隔）"`     // 端口转发重载时间（分，默认 30 分钟，多个时用-分隔）
	PortForwardReportTime string `json:"port_forward_report_time" jsonschema:"description=端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）"` // 端口转发统计上报时间（分，默认 8-12 分钟，多个时用-分隔）
}

// 设备上报IP 请求结构
type DeviceReportDeviceIpRequest struct {
	DevicePortForwardsRequest
	DeviceName string `json:"device_name,omitempty"` // 设备名称
	IPV4Addr   string `json:"ipv4_addr,omitempty"`   // IP地址V4
	IPV6Addr   string `json:"ipv6_addr,omitempty"`   // IP地址V6

	ReportType string `json:"-"` // 上报类型
}

// 设备上报IP 响应结构
type DeviceReportDeviceIpResponse struct {
	Msg string `json:"msg,omitempty"` // 消息
}

// 获取设备端口转发 请求结构
type DevicePortForwardsRequest struct {
	DeviceCode string `uri:"device_code" json:"-" binding:"required"` // 设备编码
}

// 创建端口转发统计 请求结构
type CreatePortForwardStatsRequest struct {
	DevicePortForwardsRequest
	PortForwardID int64 `uri:"port_forward_id" json:"-" binding:"required"` // 端口转发序号
	StartTime     int64 `json:"start_time" binding:"required"`              // 转发实例启动时间 (UTC)
	ActiveConn    int64 `json:"active_conn"`                                // TCP 活跃连接 / UDP session
	TotalConn     int64 `json:"total_conn"`                                 // TCP 累计连接
	BytesIn       int64 `json:"bytes_in"`                                   // 入站流量
	BytesOut      int64 `json:"bytes_out"`                                  // 出站流量
	UDPPacketsIn  int64 `json:"udp_packets_in"`                             // 入站 UDP 包数
	UDPPacketsOut int64 `json:"udp_packets_out"`                            // 出站 UDP 包数
	Errors        int64 `json:"errors"`                                     // 错误数
}
