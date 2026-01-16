namespace go open

/**
 * 开放
 */
service OpenService {
    /**
     * 获取配置
     */
    GetSettingResponse getSetting(
    /**
     * 通用头
     */
        255: required Base base
    )

    /**
     * 设备上报IP
     */
    ReportDeviceIPResponse reportDeviceIP(
    /**
     * 业务请求参数
     */
        1: required ReportDeviceIPRequest req
    /**
     * 通用头
     */
        255: required Base base
    )

    /**
     * 获取设备端口转发
     */
    GetDevicePortForwardsResponse getDevicePortForwards(
    /**
     * 业务请求参数
     */
        1: required GetDevicePortForwardsRequest req
    /**
     * 通用头
     */
        255: required Base base
    )

    /**
     * 创建端口转发统计信息
     */
    CreatePortForwardStatsResponse createPortForwardStats(
    /**
     * 业务请求参数
     */
        1: required CreatePortForwardStatsRequest req
    /**
     * 通用头
     */
        255: required Base base
    )
}

/**
 * 通用头
 */
struct Base {
    /**
     * 认证令牌
     */
    1: required string token,
    /**
     * 预留扩展
     */
    3: optional map<string,string> extra,
}

/**
 * 获取配置 响应参数
 */
struct GetSettingResponse {
    /**
     * IP上报间隔时间
     */
    1: required string IPReportIntervalTime;
    /**
     * 端口转发重载时间
     */
    2: required string PortForwardReloadTime;
    /**
     * 端口转发统计上报时间
     */
    3: required string PortForwardReportTime;
}

/**
 * 设备上报IP 请求
 */
struct ReportDeviceIPRequest {
    /**
     * 设备编码
     */
    1: required string device_code
    /**
     * 设备名称
     */
    2: optional string device_name
    /**
     * IP地址V4
     */
    3: optional string ipv4_addr
    /**
     * IP地址V6
     */
    4: optional string ipv6_addr
}

/**
 * 设备上报IP 响应
 */
struct ReportDeviceIPResponse {
    /**
     * 消息
     */
    1: optional string msg
}

/**
 * 获取设备端口转发 请求
 */
struct GetDevicePortForwardsRequest {
    /**
     * 设备编码
     */
    1: required string device_code
}

/**
 * 端口转发规则
 */
struct PortForwardRule {
    /**
     * 设备端口转发序号
     */
    1: i64 id
    /**
     * 本地监听端口
     */
    2: i32 local_port
    /**
     * 远程目标地址
     */
    3: string remote_host
    /**
     * 远程目标端口
     */
    4: i32 remote_port
    /**
     * 协议列表，支持 tcp4, tcp6, udp4, udp6
     */
    5: list<string> protocols
    /**
     * 防火墙自动放行
     */
    6: bool auto_forward
}

/**
 * 获取设备端口转发 响应
 */
struct GetDevicePortForwardsResponse {
    /**
     * 端口转发规则列表
     */
    1: optional list<PortForwardRule> rules
}

/**
 * 创建端口转发统计信息 请求
 */
struct CreatePortForwardStatsRequest {
    /**
     * 设备编码
     */
    1: required string device_code
    /**
     * 端口转发ID
     */
    2: required i64 port_forward_id
    /**
     * 转发实例启动时间 (UTC)
     */
    3: required i64 start_time
    /**
     * TCP 活跃连接 / UDP session
     */
    4: optional i64 active_conn
    /**
     * TCP 累计连接
     */
    5: optional i64 total_conn
    /**
     * 入站流量
     */
    6: optional i64 bytes_in
    /**
     * 出站流量
     */
    7: optional i64 bytes_out
    /**
     * 入站 UDP 包数
     */
    8: optional i64 udp_packets_in
    /**
     * 出站 UDP 包数
     */
    9: optional i64 udp_packets_out
    /**
     * 错误数
     */
    10: optional i64 errors
}

/**
 * 创建端口转发统计信息 响应
 */
struct CreatePortForwardStatsResponse {
}
