package schematype

const (
	RecordTypeA     string = "A"     // A记录
	RecordTypeAAAA  string = "AAAA"  // AAAA记录
	RecordTypeCNAME string = "CNAME" // CNAME记录
)

// GetRecordType 记录类型说明映射
var GetRecordTypeDescriptions = map[string]string{
	RecordTypeA:     "A记录",
	RecordTypeAAAA:  "AAAA记录",
	RecordTypeCNAME: "CNAME记录",
}

// GetRecordType 返回所有记录类型
var GetRecordTypes = []string{
	RecordTypeA,
	RecordTypeAAAA,
	RecordTypeCNAME,
}

// GetDNSRecordTypeDescriptions 记录类型说明映射
var GetDNSRecordTypeDescriptions = map[string]string{
	RecordTypeA:     "A记录",
	RecordTypeAAAA:  "AAAA记录",
	RecordTypeCNAME: "CNAME记录",
}

// GetDNSRecordType 返回 DNS 记录类型
var GetDNSRecordTypes = []string{
	RecordTypeA,
	RecordTypeAAAA,
	RecordTypeCNAME,
}

// GetRecorGetCDNRecordTypeDescriptionsdType 记录类型说明映射
var GetCDNRecordTypeDescriptions = map[string]string{
	RecordTypeA:    "A记录",
	RecordTypeAAAA: "AAAA记录",
}

// GetCDNRecordType 返回 CDN 记录类型
var GetCDNRecordTypes = []string{
	RecordTypeA,
	RecordTypeAAAA,
}

// --------

const (
	ConfigCapabilityAll string = "all" // 全部服务能力
	ConfigCapabilityDns string = "dns" // DNS服务能力
	ConfigCapabilityCdn string = "cdn" // CDN服务能力
	ConfigCapabilitySsl string = "ssl" // SSL服务能力
)

// GetConfigCapabilityDescriptions 服务能力说明映射
var GetConfigCapabilityDescriptions = map[string]string{
	ConfigCapabilityAll: "全部服务能力",
	ConfigCapabilityDns: "DNS服务能力",
	ConfigCapabilityCdn: "CDN服务能力",
	ConfigCapabilitySsl: "SSL服务能力",
}

// GetConfigCapabilitys 返回所有服务能力
var GetConfigCapabilitys = []string{
	ConfigCapabilityAll,
	ConfigCapabilityDns,
	ConfigCapabilityCdn,
	ConfigCapabilitySsl,
}

// --------

const (
	ConfigPlatformCloudflare string = "cloudflare" // Cloudflare

	ConfigPlatformTencentDns string = "tencent_dns" // 腾讯云 DNS
	ConfigPlatformTencentCdn string = "tencent_cdn" // 腾讯云 CDN
	ConfigPlatformTencentEO  string = "tencent_eo"  // 腾讯云 EdgeOne

	ConfigPlatformAliyunDns string = "aliyun_dns" // 阿里云 DNS
	ConfigPlatformAliyunCdn string = "aliyun_cdn" // 阿里云 CDN
	ConfigPlatformAliyunEsa string = "aliyun_esa" // 阿里云 ESA

	ConfigPlatformBaiduDns string = "baidu_dns" // 百度云 DNS

	ConfigPlatformHuaweiDns string = "huawei_dns" // 华为云 DNS
	ConfigPlatformHuaweiCdn string = "huawei_cdn" // 华为云 CDN

	ConfigPlatformvolcengineDns string = "volcengine_dns" // 火山引擎 DNS
)

// GetConfigPlatformDescriptions 平台类型说明映射
var GetConfigPlatformDescriptions = map[string]string{
	ConfigPlatformCloudflare:    "Cloudflare",
	ConfigPlatformTencentDns:    "腾讯云 DNS",
	ConfigPlatformTencentCdn:    "腾讯云 CDN",
	ConfigPlatformTencentEO:     "腾讯云 EdgeOne",
	ConfigPlatformAliyunDns:     "阿里云 DNS",
	ConfigPlatformAliyunCdn:     "阿里云 CDN",
	ConfigPlatformAliyunEsa:     "阿里云 ESA",
	ConfigPlatformBaiduDns:      "百度云 DNS",
	ConfigPlatformHuaweiDns:     "华为云 DNS",
	ConfigPlatformHuaweiCdn:     "华为云 CDN",
	ConfigPlatformvolcengineDns: "火山引擎 DNS",
}

// GetConfigPlatforms 返回所有平台类型
var GetConfigPlatforms = []string{
	ConfigPlatformCloudflare,
	ConfigPlatformTencentDns,
	ConfigPlatformTencentCdn,
	ConfigPlatformTencentEO,
	ConfigPlatformAliyunDns,
	ConfigPlatformAliyunCdn,
	ConfigPlatformAliyunEsa,
	ConfigPlatformBaiduDns,
	ConfigPlatformHuaweiDns,
	ConfigPlatformHuaweiCdn,
	ConfigPlatformvolcengineDns,
}

// --------

const (
	WebhookPlatformWechatWork string = "wechat_work" // 企业微信
	WebhookPlatformDingtalk   string = "dingtalk"    // 钉钉
	WebhookPlatformFeishu     string = "feishu"      // 飞书
	WebhookPlatformBark       string = "bark"        // Bark
	WebhookPlatformFt07       string = "ft07"        // Server酱
	WebhookPlatformPushDeer   string = "push_deer"   // PushDeer
	WebhookPlatformCustomHTTP string = "custom_http" // 自定义 HTTP 平台
)

// GetWebhookPlatformDescriptions Webhook 平台类型说明映射
var GetWebhookPlatformDescriptions = map[string]string{
	WebhookPlatformWechatWork: "企业微信",
	WebhookPlatformDingtalk:   "钉钉",
	WebhookPlatformFeishu:     "飞书",
	WebhookPlatformBark:       "Bark",
	WebhookPlatformFt07:       "Server酱",
	WebhookPlatformPushDeer:   "PushDeer",
	WebhookPlatformCustomHTTP: "自定义 HTTP 平台",
}

// GetWebhookPlatforms 返回所有 Webhook 平台
var GetWebhookPlatforms = []string{
	WebhookPlatformWechatWork,
	WebhookPlatformDingtalk,
	WebhookPlatformFeishu,
	WebhookPlatformBark,
	WebhookPlatformFt07,
	WebhookPlatformPushDeer,
	WebhookPlatformCustomHTTP,
}

// --------

const (
	WebhookSubscriptionObjTypeCDNRecord     string = "cdn_record"     // CDN记录
	WebhookSubscriptionObjTypeDNSRecord     string = "dns_record"     // DNS记录
	WebhookSubscriptionObjTypeDevice        string = "device"         // 设备
	WebhookSubscriptionObjTypeIPGroup       string = "ip_group"       // IP组
	WebhookSubscriptionObjTypeDomainMonitor string = "domain_monitor" // 域名监控
	WebhookSubscriptionObjTypeSSLRecord     string = "ssl_record"     // SSL记录
)

// GetWebhookSubscriptionObjTypeDescriptions Webhook 订阅对象类型说明映射
var GetWebhookSubscriptionObjTypeDescriptions = map[string]string{
	WebhookSubscriptionObjTypeCDNRecord:     "CDN记录",
	WebhookSubscriptionObjTypeDNSRecord:     "DNS记录",
	WebhookSubscriptionObjTypeDevice:        "设备",
	WebhookSubscriptionObjTypeIPGroup:       "IP组",
	WebhookSubscriptionObjTypeDomainMonitor: "域名监控",
	WebhookSubscriptionObjTypeSSLRecord:     "SSL记录",
}

// GetWebhookSubscriptionObjTypes 返回所有 Webhook 订阅对象类型
var GetWebhookSubscriptionObjTypes = []string{
	WebhookSubscriptionObjTypeCDNRecord,
	WebhookSubscriptionObjTypeDNSRecord,
	WebhookSubscriptionObjTypeDevice,
	WebhookSubscriptionObjTypeIPGroup,
	WebhookSubscriptionObjTypeDomainMonitor,
	WebhookSubscriptionObjTypeSSLRecord,
}

// --------

const (
	DeviceIpReportReportTypeUnknown string = "unknown" // 未知
	DeviceIpReportReportTypeAPI     string = "api"     // API
	DeviceIpReportReportTypeGrpc    string = "grpc"    // gRPC
	DeviceIpReportReportTypeThrift  string = "thrift"  // Thrift
	DeviceIpReportReportTypeDubbo   string = "dubbo"   // Dubbo
	DeviceIpReportReportTypeTars    string = "tars"    // Tars
)

// GetDeviceIpReportReportTypeDescriptions 设备 IP 上报方式说明映射
var GetDeviceIpReportReportTypeDescriptions = map[string]string{
	DeviceIpReportReportTypeUnknown: "未知",
	DeviceIpReportReportTypeAPI:     "API",
	DeviceIpReportReportTypeGrpc:    "gRPC",
	DeviceIpReportReportTypeThrift:  "Thrift",
	DeviceIpReportReportTypeDubbo:   "Dubbo",
	DeviceIpReportReportTypeTars:    "Tars",
}

// GetDeviceIpReportReportTypes 返回所有设备 IP 上报方式
var GetDeviceIpReportReportTypes = []string{
	DeviceIpReportReportTypeUnknown,
	DeviceIpReportReportTypeAPI,
	DeviceIpReportReportTypeGrpc,
	DeviceIpReportReportTypeThrift,
	DeviceIpReportReportTypeDubbo,
	DeviceIpReportReportTypeTars,
}

// --------

const (
	DomainMonitorHistoryCheckTypeUnknown  string = "unknown"  // 未知
	DomainMonitorHistoryCheckTypeStandard string = "standard" // 标准
	DomainMonitorHistoryCheckTypeResty    string = "resty"    // Resty
	DomainMonitorHistoryCheckTypeReq      string = "req"      // Req
)

// GetDomainMonitorHistoryCheckTypeDescriptions 域名监控检查方式说明映射
var GetDomainMonitorHistoryCheckTypeDescriptions = map[string]string{
	DomainMonitorHistoryCheckTypeUnknown:  "未知",
	DomainMonitorHistoryCheckTypeStandard: "标准",
	DomainMonitorHistoryCheckTypeResty:    "Resty",
	DomainMonitorHistoryCheckTypeReq:      "Req",
}

// GetDomainMonitorHistoryCheckTypes 返回所有域名监控检查方式
var GetDomainMonitorHistoryCheckTypes = []string{
	DomainMonitorHistoryCheckTypeUnknown,
	DomainMonitorHistoryCheckTypeStandard,
	DomainMonitorHistoryCheckTypeResty,
	DomainMonitorHistoryCheckTypeReq,
}

// --------

const (
	LogSourceUnknown string = "unknown" // 未知来源
	LogSourceSystem  string = "system"  // 系统自动
	LogSourceTask    string = "task"    // 定时任务
	LogSourceUser    string = "user"    // 用户手动
	LogSourceGin     string = "gin"     // Gin
	LogSourceAPI     string = "api"     // API
	LogSourceMCP     string = "mcp"     // MCP
	LogSourceGrpc    string = "grpc"    // gRPC
	LogSourceThrift  string = "thrift"  // Thrift
	LogSourceDubbo   string = "dubbo"   // Dubbo
	LogSourceTars    string = "tars"    // Tars
)

// GetLogSourceDescriptions 操作来源说明映射
var GetLogSourceDescriptions = map[string]string{
	LogSourceUnknown: "未知来源",
	LogSourceSystem:  "系统自动",
	LogSourceTask:    "定时任务",
	LogSourceUser:    "用户手动",
	LogSourceGin:     "Gin",
	LogSourceAPI:     "API",
	LogSourceMCP:     "MCP",
	LogSourceGrpc:    "gRPC",
	LogSourceThrift:  "Thrift",
	LogSourceDubbo:   "Dubbo",
	LogSourceTars:    "Tars",
}

// GetLogSources 返回所有操作来源
var GetLogSources = []string{
	LogSourceUnknown,
	LogSourceSystem,
	LogSourceTask,
	LogSourceUser,
	LogSourceGin,
	LogSourceAPI,
	LogSourceMCP,
	LogSourceGrpc,
	LogSourceThrift,
	LogSourceDubbo,
	LogSourceTars,
}

const (
	LogOpTypeUnknown      string = "unknown"       // 未知操作
	LogOpTypeDataCreate   string = "data_create"   // 数据创建
	LogOpTypeDataUpdate   string = "data_update"   // 数据更新
	LogOpTypeDataDelete   string = "data_delete"   // 数据删除
	LogOpTypeDeviceReport string = "device_report" // 设备上报
	LogOpTypeIPRefresh    string = "ip_refresh"    // IP刷新
)

// GetLogOpTypeDescriptions 操作类型说明映射
var GetLogOpTypeDescriptions = map[string]string{
	LogOpTypeUnknown:      "未知操作",
	LogOpTypeDataCreate:   "数据创建",
	LogOpTypeDataUpdate:   "数据更新",
	LogOpTypeDataDelete:   "数据删除",
	LogOpTypeDeviceReport: "设备上报",
	LogOpTypeIPRefresh:    "IP刷新",
}

// GetLogOpTypes 返回所有操作类型
var GetLogOpTypes = []string{
	LogOpTypeUnknown,
	LogOpTypeDataCreate,
	LogOpTypeDataUpdate,
	LogOpTypeDataDelete,
	LogOpTypeDeviceReport,
	LogOpTypeIPRefresh,
}

// --------

const (
	LogObjTypeUnknown string = "unknown" // 未知对象

	LogObjTypeCache               string = "cache"                // 缓存
	LogObjTypeCDNRecord           string = "cdn_record"           // CDN记录
	LogObjTypeDNSRecord           string = "dns_record"           // DNS记录
	LogObjTypeDevice              string = "device"               // 设备
	LogObjTypeDevicePortForward   string = "device_port_forward"  // 端口转发
	LogObjTypeDomainMonitor       string = "domain_monitor"       // 域名监控
	LogObjTypeIPGroup             string = "ip_group"             // IP组
	LogObjTypeIPGroupDevice       string = "ip_group_device"      // IP组设备
	LogObjTypePlatformConfig      string = "platform_config"      // 平台配置
	LogObjTypeSSLRecord           string = "ssl_record"           // SSL记录
	LogObjTypeWebhook             string = "webhook"              // 通知
	LogObjTypeWebhookSubscription string = "webhook_subscription" // 通知订阅

	LogObjTypeDeviceIPRecord        string = "device_ip_report"        // 设备IP记录
	LogObjTypeDomainMonitorHistory  string = "domain_monitor_history"  // 域名监控历史记录
	LogObjTypeHttpLog               string = "http_log"                // http请求日志
	LogObjTypeOperationLog          string = "operation_log"           // 操作日志
	LogObjTypePortForwardStats      string = "port_forward_stats"      // 端口转发统计快照记录
	LogObjTypeRestyLog              string = "resty_log"               // resty请求日志
	LogObjTypeSSLCertificateHistory string = "ssl_certificate_history" // SSL证书申请历史记录
	LogObjTypeWebhookHistory        string = "webhook_history"         // 通知历史记录
)

// GetLogObjTypeDescriptions 对象类型说明映射
var GetLogObjTypeDescriptions = map[string]string{
	LogObjTypeUnknown: "未知对象",

	LogObjTypeCache:               "缓存",
	LogObjTypeCDNRecord:           "CDN记录",
	LogObjTypeDNSRecord:           "DNS记录",
	LogObjTypeDevice:              "设备",
	LogObjTypeDevicePortForward:   "端口转发",
	LogObjTypeDomainMonitor:       "域名监控",
	LogObjTypeIPGroup:             "IP组",
	LogObjTypeIPGroupDevice:       "IP组设备",
	LogObjTypePlatformConfig:      "平台配置",
	LogObjTypeSSLRecord:           "SSL记录",
	LogObjTypeWebhook:             "通知",
	LogObjTypeWebhookSubscription: "通知订阅",

	LogObjTypeDeviceIPRecord:        "设备IP记录",
	LogObjTypeDomainMonitorHistory:  "域名监控历史记录",
	LogObjTypeHttpLog:               "http请求日志",
	LogObjTypeOperationLog:          "操作日志",
	LogObjTypePortForwardStats:      "端口转发统计快照记录",
	LogObjTypeRestyLog:              "resty请求日志",
	LogObjTypeSSLCertificateHistory: "SSL证书申请历史记录",
	LogObjTypeWebhookHistory:        "通知历史记录",
}

// GetLogObjTypes 返回所有对象类型
var GetLogObjTypes = []string{
	LogObjTypeUnknown,

	LogObjTypeCache,
	LogObjTypeCDNRecord,
	LogObjTypeDNSRecord,
	LogObjTypeDevice,
	LogObjTypeDevicePortForward,
	LogObjTypeDomainMonitor,
	LogObjTypeIPGroup,
	LogObjTypeIPGroupDevice,
	LogObjTypePlatformConfig,
	LogObjTypeSSLRecord,
	LogObjTypeWebhook,
	LogObjTypeWebhookSubscription,

	LogObjTypeDeviceIPRecord,
	LogObjTypeDomainMonitorHistory,
	LogObjTypeHttpLog,
	LogObjTypeOperationLog,
	LogObjTypePortForwardStats,
	LogObjTypeRestyLog,
	LogObjTypeSSLCertificateHistory,
	LogObjTypeWebhookHistory,
}

// --------

const (
	LogStatusUnknown string = "unknown" // 未知状态
	LogStatusSuccess string = "success" // 成功
	LogStatusFailed  string = "failed"  // 失败
)

// GetLogStatusDescriptions 状态说明映射
var GetLogStatusDescriptions = map[string]string{
	LogStatusUnknown: "未知状态",
	LogStatusSuccess: "成功",
	LogStatusFailed:  "失败",
}

// GetLogStatuss 返回所有状态
var GetLogStatuss = []string{
	LogStatusUnknown,
	LogStatusSuccess,
	LogStatusFailed,
}

// --------

const (
	DomainMonitorCheckProtocolHTTP  string = "http"  // HTTP 协议
	DomainMonitorCheckProtocolHTTPS string = "https" // HTTPS 协议
)

var GetDomainMonitorCheckProtocolDescriptions = map[string]string{
	DomainMonitorCheckProtocolHTTP:  "HTTP协议",
	DomainMonitorCheckProtocolHTTPS: "HTTPS协议",
}

// GetDomainMonitorCheckProtocols 返回所有检查协议
var GetDomainMonitorCheckProtocols = []string{
	DomainMonitorCheckProtocolHTTP,
	DomainMonitorCheckProtocolHTTPS,
}

// --------

const (
	DomainMonitorStatusSuccess string = "success" // 成功运行
	DomainMonitorStatusFailure string = "failure" // 检查失败
	DomainMonitorStatusUnknown string = "unknown" // 未知状态
)

// GetDomainMonitorStatusDescriptions 运行状态说明映射
var GetDomainMonitorStatusDescriptions = map[string]string{
	DomainMonitorStatusSuccess: "成功运行",
	DomainMonitorStatusFailure: "检查失败",
	DomainMonitorStatusUnknown: "未知状态",
}

// GetDomainMonitorStatuss 返回所有运行状态
var GetDomainMonitorStatuss = []string{
	DomainMonitorStatusSuccess,
	DomainMonitorStatusFailure,
	DomainMonitorStatusUnknown,
}

// --------

const (
	DevicePortForwardProtocolTCP4 string = "tcp4" // TCP4 协议
	DevicePortForwardProtocolTCP6 string = "tcp6" // TCP6 协议
	DevicePortForwardProtocolUDP4 string = "udp4" // UDP4 协议
	DevicePortForwardProtocolUDP6 string = "udp6" // UDP6 协议
)

// GetDevicePortForwardProtocolDescriptions 端口转发协议说明映射
var GetDevicePortForwardProtocolDescriptions = map[string]string{
	DevicePortForwardProtocolTCP4: "TCP4协议",
	DevicePortForwardProtocolTCP6: "TCP6协议",
	DevicePortForwardProtocolUDP4: "UDP4协议",
	DevicePortForwardProtocolUDP6: "UDP6协议",
}

// GetDevicePortForwardProtocols 返回所有端口转发协议
var GetDevicePortForwardProtocols = []string{
	DevicePortForwardProtocolTCP4,
	DevicePortForwardProtocolTCP6,
	DevicePortForwardProtocolUDP4,
	DevicePortForwardProtocolUDP6,
}

const (
	SSLRecordStatusValid   string = "valid"   // 有效
	SSLRecordStatusExpired string = "expired" // 过期
	SSLRecordStatusRevoked string = "revoked" // 吊销
	SSLRecordStatusPending string = "pending" // 待处理
)

// GetSSLRecordStatusDescriptions SSL记录状态说明映射
var GetSSLRecordStatusDescriptions = map[string]string{
	SSLRecordStatusValid:   "有效",
	SSLRecordStatusExpired: "过期",
	SSLRecordStatusRevoked: "吊销",
	SSLRecordStatusPending: "待处理",
}

// GetSSLRecordStatuss 返回所有SSL记录状态
var GetSSLRecordStatuss = []string{
	SSLRecordStatusValid,
	SSLRecordStatusExpired,
	SSLRecordStatusRevoked,
	SSLRecordStatusPending,
}

const (
	SSLRecordProviderLetsencrypt string = "letsencrypt" // Letsencrypt 提供商
)

// GetSSLRecordProviderDescriptions SSL记录提供商说明映射
var GetSSLRecordProviderDescriptions = map[string]string{
	SSLRecordProviderLetsencrypt: "Letsencrypt提供商",
}

// GetSSLRecordProviders 返回所有SSL记录提供商
var GetSSLRecordProviders = []string{
	SSLRecordProviderLetsencrypt,
}
