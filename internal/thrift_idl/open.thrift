namespace go open

/**
 * 开放接口
 */
service OpenService {
    /**
     * 设备上报 IP
     */
    ReportDeviceIPResponse reportDeviceIP(
        1: required ReportDeviceIPRequest req
    )

    /**
     * 获取设备端口转发
     */
    GetDevicePortForwardsResponse getDevicePortForwards(
        1: required GetDevicePortForwardsRequest req
    )

    /**
     * 创建端口转发统计信息
     */
    CreatePortForwardStatsResponse createPortForwardStats(
        1: required CreatePortForwardStatsRequest req
    )
}

/**
 * 设备上报 IP 请求
 */
struct ReportDeviceIPRequest {
    1: required string device_code // 设备编码
    2: optional string device_name // 设备名称
    3: optional string ipv4_addr   // IP地址V4
    4: optional string ipv6_addr   // IP地址V6
}

/**
 * 设备上报 IP 响应
 */
struct ReportDeviceIPResponse {
    1: optional string msg // 消息
}

/**
 * 获取设备端口转发 请求
 */
struct GetDevicePortForwardsRequest {
    1: required string device_code // 设备编码
}

/**
 * 端口转发规则
 */
struct PortForwardRule {
    1: i64          id           // 设备端口转发序号
    2: i32          local_port   // 本地监听端口
    3: string       remote_host  // 远程目标地址
    4: i32          remote_port  // 远程目标端口
    5: list<string> protocols    // 协议列表，支持 tcp4, tcp6, udp4, udp6
    6: bool         auto_forward // 防火墙自动放行
}

/**
 * 获取设备端口转发 响应
 */
struct GetDevicePortForwardsResponse {
    1: optional list<PortForwardRule> rules // 端口转发规则列表
}

/**
 * 创建端口转发统计信息 请求
 */
struct CreatePortForwardStatsRequest {
    1:  required string device_code     // 设备编码
    2:  required i64    port_forward_id // 端口转发ID
    3:  required i64    start_time      // 转发实例启动时间 (UTC)
    4:  optional i64    active_conn     // TCP 活跃连接 / UDP session
    5:  optional i64    total_conn      // TCP 累计连接
    6:  optional i64    bytes_in        // 入站流量
    7:  optional i64    bytes_out       // 出站流量
    8:  optional i64    udp_packets_in  // 入站 UDP 包数
    9:  optional i64    udp_packets_out // 出站 UDP 包数
    10: optional i64    errors          // 错误数
}

/**
 * 创建端口转发统计信息 响应
 */
struct CreatePortForwardStatsResponse {
}
