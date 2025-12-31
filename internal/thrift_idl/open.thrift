// Thrift IDL 文件：OpenService
// 基于 gRPC proto 转换，移除 HTTP 注解和验证规则（Thrift 无内置支持）
// 生成 Go 代码：thrift -r --gen go:package=internal/gen-thrift OpenService.thrift
// 生成 Go 代码：thrift -r --gen go:package=thrift_generated --out . ./thrift_idl/*.thrift

// include "shared.thrift"  // 可选：如果有共享类型

// Go 包路径
namespace go open

// 异常定义（Thrift 标准）
exception TException {
    1: i32    errorCode,
    2: string message
}

service OpenService {
    // 设备上报IP（返回消息字符串）
    string ReportDeviceIP(1: ReportDeviceIPRequest req) throws (1: TException ouch),

    // 获取设备端口转发（返回规则列表）
    list<PortForwardRule> GetDevicePortForwards(1: GetDevicePortForwardsRequest req) throws (1: TException ouch),

    // 创建端口转发统计信息（返回 void，成功无响应）
    void CreatePortForwardStats(1: CreatePortForwardStatsRequest req) throws (1: TException ouch)
}

// 设备上报IP 请求参数
struct ReportDeviceIPRequest {
    1: required string device_code,  // 设备编码（required 以匹配 proto 的 min_len:1）
    2: optional string device_name,  // 设备名称（可选）
    3: optional string ipv4_addr,    // IP地址V4（可选）
    4: optional string ipv6_addr     // IP地址V6（可选）
}

// 设备上报IP 响应（简化：Thrift 方法直接返回 string，无需 struct）
/*
 * struct ReportDeviceIPResponse {
 * 1: string msg  // 如果需要复杂响应，可用此；否则方法返回 string
 * }
 */

// 获取设备端口转发 请求参数
struct GetDevicePortForwardsRequest {
    1: required string device_code // 设备编码
}

// 端口转发规则
struct PortForwardRule {
    1: i64          id,           // 设备端口转发序号
    2: i32          local_port,   // 本地监听端口
    3: string       remote_host,  // 远程目标地址
    4: i32          remote_port,  // 远程目标端口
    5: list<string> protocols,    // 协议列表
    6: bool         auto_forward  // 防火墙自动放行
}

// 获取设备端口转发 响应（直接用 list<PortForwardRule> 返回，无需 struct）
/*
 * struct GetDevicePortForwardsResponse {
 * 1: list<PortForwardRule> rules
 * }
 */

// 创建端口转发统计信息 请求参数
struct CreatePortForwardStatsRequest {
    1:  required string device_code,      // 设备编码
    2:  required i64    port_forward_id,  // 端口转发ID
    3:  required i64    start_time,       // 转发实例启动时间 (UTC)
    4:           i64    active_conn,      // TCP 活跃连接 / UDP session
    5:           i64    total_conn,       // TCP 累计连接
    6:           i64    bytes_in,         // 入站流量
    7:           i64    bytes_out,        // 出站流量
    8:           i64    udp_packets_in,   // 入站 UDP 包数
    9:           i64    udp_packets_out,  // 出站 UDP 包数
    10:          i64    errors            // 错误数
}

// 创建端口转发统计信息 响应（void，无 struct）
/*
 * struct CreatePortForwardStatsResponse {
 * // 空响应
 * }
 */