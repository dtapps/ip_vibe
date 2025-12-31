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
    void createPortForwardStats(
        1: required CreatePortForwardStatsRequest req
    )
}

/**
 * 设备上报 IP 请求
 */
struct ReportDeviceIPRequest {
    1: required string device_code // 设备编码
    2: optional string device_name // 设备名称
    3: optional string ipv4_addr   // IPv4
    4: optional string ipv6_addr   // IPv6
}

/**
 * 设备上报 IP 响应
 */
struct ReportDeviceIPResponse {
    1: optional string msg
}

/**
 * 获取设备端口转发 请求
 */
struct GetDevicePortForwardsRequest {
    1: required string device_code
}

/**
 * 端口转发规则
 */
struct PortForwardRule {
    1: i64          id
    2: i32          local_port
    3: string       remote_host
    4: i32          remote_port
    5: list<string> protocols // tcp4, tcp6, udp4, udp6
    6: bool         auto_forward
}

/**
 * 获取设备端口转发 响应
 */
struct GetDevicePortForwardsResponse {
    1: list<PortForwardRule> rules
}

/**
 * 创建端口转发统计信息 请求
 */
struct CreatePortForwardStatsRequest {
    1:  required string device_code
    2:           i64    port_forward_id
    3:           i64    start_time
    4:           i64    active_conn
    5:           i64    total_conn
    6:           i64    bytes_in
    7:           i64    bytes_out
    8:           i64    udp_packets_in
    9:           i64    udp_packets_out
    10:          i64    errors
}

/**
 * 创建端口转发统计信息 响应
 */
struct CreatePortForwardStatsResponse {
}
