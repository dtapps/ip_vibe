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
	ConfigCapabilityTun string = "tun" // TUN服务能力
)

// GetConfigCapabilityDescriptions 服务能力说明映射
var GetConfigCapabilityDescriptions = map[string]string{
	ConfigCapabilityAll: "全部服务能力",
	ConfigCapabilityDns: "DNS服务能力",
	ConfigCapabilityCdn: "CDN服务能力",
	ConfigCapabilitySsl: "SSL服务能力",
	ConfigCapabilityTun: "TUN服务能力",
}

// GetConfigCapabilitys 返回所有服务能力
var GetConfigCapabilitys = []string{
	ConfigCapabilityAll,
	ConfigCapabilityDns,
	ConfigCapabilityCdn,
	ConfigCapabilitySsl,
	ConfigCapabilityTun,
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
	ConfigPlatformBaiduCdn string = "baidu_cdn" // 百度云 CDN

	ConfigPlatformHuaweiDns string = "huawei_dns" // 华为云 DNS
	ConfigPlatformHuaweiCdn string = "huawei_cdn" // 华为云 CDN

	ConfigPlatformVolcengineDns string = "volcengine_dns" // 火山引擎 DNS

	ConfigPlatformJdcloudCdn string = "jdcloud_cdn" // 京东云 CDN

	ConfigPlatformKsyunCdn string = "ksyun_cdn" // 金山云 CDN

	ConfigPlatformQiniuCdn string = "qiniu_cdn" // 七牛云 CDN

	ConfigPlatformUpyunCdn string = "upyun_cdn" // 又拍云 CDN
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
	ConfigPlatformBaiduCdn:      "百度云 CDN",
	ConfigPlatformHuaweiDns:     "华为云 DNS",
	ConfigPlatformHuaweiCdn:     "华为云 CDN",
	ConfigPlatformVolcengineDns: "火山引擎 DNS",
	ConfigPlatformJdcloudCdn:    "京东云 CDN",
	ConfigPlatformKsyunCdn:      "金山云 CDN",
	ConfigPlatformQiniuCdn:      "七牛云 CDN",
	ConfigPlatformUpyunCdn:      "又拍云 CDN",
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
	ConfigPlatformBaiduCdn,
	ConfigPlatformHuaweiDns,
	ConfigPlatformHuaweiCdn,
	ConfigPlatformVolcengineDns,
	ConfigPlatformJdcloudCdn,
	ConfigPlatformKsyunCdn,
	ConfigPlatformQiniuCdn,
	ConfigPlatformUpyunCdn,
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
	WebhookSubscriptionObjTypeDeviceGroup   string = "device_group"   // 设备组
	WebhookSubscriptionObjTypeIPGroup       string = "ip_group"       // IP组
	WebhookSubscriptionObjTypeDomainMonitor string = "domain_monitor" // 域名监控
	WebhookSubscriptionObjTypeSSLRecord     string = "ssl_record"     // SSL记录
	WebhookSubscriptionObjTypeTUNRecord     string = "tun_record"     // TUN记录
)

// GetWebhookSubscriptionObjTypeDescriptions Webhook 订阅对象类型说明映射
var GetWebhookSubscriptionObjTypeDescriptions = map[string]string{
	WebhookSubscriptionObjTypeCDNRecord:     "CDN记录",
	WebhookSubscriptionObjTypeDNSRecord:     "DNS记录",
	WebhookSubscriptionObjTypeDevice:        "设备",
	WebhookSubscriptionObjTypeDeviceGroup:   "设备组",
	WebhookSubscriptionObjTypeIPGroup:       "IP组",
	WebhookSubscriptionObjTypeDomainMonitor: "域名监控",
	WebhookSubscriptionObjTypeSSLRecord:     "SSL记录",
	WebhookSubscriptionObjTypeTUNRecord:     "TUN记录",
}

// GetWebhookSubscriptionObjTypes 返回所有 Webhook 订阅对象类型
var GetWebhookSubscriptionObjTypes = []string{
	WebhookSubscriptionObjTypeCDNRecord,
	WebhookSubscriptionObjTypeDNSRecord,
	WebhookSubscriptionObjTypeDevice,
	WebhookSubscriptionObjTypeDeviceGroup,
	WebhookSubscriptionObjTypeIPGroup,
	WebhookSubscriptionObjTypeDomainMonitor,
	WebhookSubscriptionObjTypeSSLRecord,
	WebhookSubscriptionObjTypeTUNRecord,
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
	LogSourceUnknown string = "unknown" // 未知
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
	LogSourceUnknown: "未知",
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
	LogOpTypeUnknown string = "Unknown" // 未知

	LogOpTypeOpCreate    string = "OpCreate"    // 创建单条
	LogOpTypeOpUpdate    string = "OpUpdate"    // 更新多条
	LogOpTypeOpUpdateOne string = "OpUpdateOne" // 更新单条
	LogOpTypeOpDelete    string = "OpDelete"    // 删除多条
	LogOpTypeOpDeleteOne string = "OpDeleteOne" // 删除单条
)

// GetLogOpTypeDescriptions 操作类型说明映射
var GetLogOpTypeDescriptions = map[string]string{
	LogOpTypeUnknown: "未知",

	LogOpTypeOpCreate:    "创建单条",
	LogOpTypeOpUpdate:    "更新多条",
	LogOpTypeOpUpdateOne: "更新单条",
	LogOpTypeOpDelete:    "删除多条",
	LogOpTypeOpDeleteOne: "删除单条",
}

// GetLogOpTypes 返回所有操作类型
var GetLogOpTypes = []string{
	LogOpTypeUnknown,

	LogOpTypeOpCreate,
	LogOpTypeOpUpdate,
	LogOpTypeOpUpdateOne,
	LogOpTypeOpDelete,
	LogOpTypeOpDeleteOne,
}

// --------

const (
	LogObjTypeUnknown string = "Unknown" // 未知

	// ent/mutation.go

	LogObjTypeCDNRecord                      string = "CDNRecord"                     // CDN记录
	LogObjTypeCache                          string = "Cache"                         // 缓存
	LogObjTypeCloudflareTunnel               string = "CloudflareTunnel"              // Cloudflare隧道
	LogObjTypeCloudflareTunnelIngress        string = "CloudflareTunnelIngress"       // Cloudflare隧道入站规则
	LogObjTypeCloudflareZone                 string = "CloudflareZone"                // Cloudflare区域
	LogObjTypeCloudflareZoneRecord           string = "CloudflareZoneRecord"          // Cloudflare区域记录
	LogObjTypeDNSRecord                      string = "DNSRecord"                     // DNS记录
	LogObjTypeCronJob                        string = "CronJob"                       // 定时任务
	LogObjTypeDevice                         string = "Device"                        // 设备
	LogObjTypeDeviceGroup                    string = "DeviceGroup"                   // 设备组
	LogObjTypeDeviceGroupDevice              string = "DeviceGroupDevice"             // 设备组设备
	LogObjTypeDevicePortForward              string = "DevicePortForward"             // 端口转发
	LogObjTypeDomainMonitor                  string = "DomainMonitor"                 // 域名监控
	LogObjTypeEdgeOneZone                    string = "EdgeOneZone"                   // EdgeOne站点
	LogObjTypeEdgeOneZoneAccelerationDomain  string = "EdgeOneZoneAccelerationDomain" // EdgeOne加速域名
	LogObjTypeEdgeOneZoneDNSRecord           string = "EdgeOneZoneDNSRecord"          // EdgeOneDNS记录
	LogObjTypeEdgeOneZoneOriginGroup         string = "EdgeOneZoneOriginGroup"        // EdgeOne站点源站组
	LogObjTypeTypeEsaSite                    string = "EsaSite"                       // Esa站点
	LogObjTypeTypeEsaSiteOriginPool          string = "EsaSiteOriginPool"             // Esa站点源站池
	LogObjTypeTypeEsaSiteEdgeContainerRecord string = "EsaSiteEdgeContainerRecord"    // Esa站点边缘容器记录
	LogObjTypeTypeEsaSiteEdgeRoutineRecord   string = "EsaSiteEdgeRoutineRecord"      // Esa站点边缘路由记录
	LogObjTypeTypeEsaSiteRecord              string = "EsaSiteRecord"                 // Esa站点记录
	LogObjTypeIpGroup                        string = "IpGroup"                       // IP组
	LogObjTypeIpGroupDevice                  string = "IpGroupDevice"                 // IP组设备
	LogObjTypePlatformConfig                 string = "PlatformConfig"                // 平台配置
	LogObjTypeSSLImport                      string = "SSLImport"                     // SSL导入
	LogObjTypeSSLRecord                      string = "SSLRecord"                     // SSL记录
	LogObjTypeTUNRecord                      string = "TUNRecord"                     // TUN记录
	LogObjTypeWebhook                        string = "Webhook"                       // 通知
	LogObjTypeWebhookSubscription            string = "WebhookSubscription"           // 通知订阅

	// ent_log/mutation.go

	LogObjTypeDeviceIpReport        string = "DeviceIpReport"        // 设备IP记录
	LogObjTypeDomainMonitorHistory  string = "DomainMonitorHistory"  // 域名监控历史记录
	LogObjTypeHttpLog               string = "HttpLog"               // http请求日志
	LogObjTypeOperationLog          string = "OperationLog"          // 操作日志
	LogObjTypePortForwardStats      string = "PortForwardStats"      // 端口转发统计快照记录
	LogObjTypeReqLog                string = "ReqLog"                // req请求日志
	LogObjTypeRestyLog              string = "RestyLog"              // resty请求日志
	LogObjTypeSSLCertificateHistory string = "SSLCertificateHistory" // SSL证书申请历史记录
	LogObjTypeWebhookHistory        string = "WebhookHistory"        // 通知历史记录
)

// GetLogObjTypeDescriptions 对象类型说明映射
var GetLogObjTypeDescriptions = map[string]string{
	LogObjTypeUnknown: "未知",

	LogObjTypeCache:                          "缓存",
	LogObjTypeCloudflareTunnel:               "Cloudflare隧道",
	LogObjTypeCloudflareTunnelIngress:        "Cloudflare隧道入站规则",
	LogObjTypeCloudflareZone:                 "Cloudflare区域",
	LogObjTypeCloudflareZoneRecord:           "Cloudflare区域记录",
	LogObjTypeCDNRecord:                      "CDN记录",
	LogObjTypeDNSRecord:                      "DNS记录",
	LogObjTypeCronJob:                        "定时任务",
	LogObjTypeDevice:                         "设备",
	LogObjTypeDeviceGroup:                    "设备组",
	LogObjTypeDeviceGroupDevice:              "设备组设备",
	LogObjTypeDevicePortForward:              "端口转发",
	LogObjTypeDomainMonitor:                  "域名监控",
	LogObjTypeEdgeOneZone:                    "EdgeOne站点",
	LogObjTypeEdgeOneZoneAccelerationDomain:  "EdgeOne加速域名",
	LogObjTypeEdgeOneZoneDNSRecord:           "EdgeOneDNS记录",
	LogObjTypeEdgeOneZoneOriginGroup:         "EdgeOne站点源站组",
	LogObjTypeTypeEsaSite:                    "Esa站点",
	LogObjTypeTypeEsaSiteOriginPool:          "Esa站点源站池",
	LogObjTypeTypeEsaSiteEdgeContainerRecord: "Esa站点边缘容器记录",
	LogObjTypeTypeEsaSiteEdgeRoutineRecord:   "Esa站点边缘路由记录",
	LogObjTypeTypeEsaSiteRecord:              "Esa站点记录",
	LogObjTypeIpGroup:                        "IP组",
	LogObjTypeIpGroupDevice:                  "IP组设备",
	LogObjTypePlatformConfig:                 "平台配置",
	LogObjTypeSSLImport:                      "SSL导入",
	LogObjTypeSSLRecord:                      "SSL记录",
	LogObjTypeTUNRecord:                      "TUN记录",
	LogObjTypeWebhook:                        "通知",
	LogObjTypeWebhookSubscription:            "通知订阅",

	LogObjTypeDeviceIpReport:        "设备IP记录",
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
	LogObjTypeCloudflareTunnel,
	LogObjTypeCloudflareTunnelIngress,
	LogObjTypeCloudflareZone,
	LogObjTypeCloudflareZoneRecord,
	LogObjTypeCDNRecord,
	LogObjTypeDNSRecord,
	LogObjTypeCronJob,
	LogObjTypeDevice,
	LogObjTypeDeviceGroup,
	LogObjTypeDeviceGroupDevice,
	LogObjTypeDevicePortForward,
	LogObjTypeDomainMonitor,
	LogObjTypeEdgeOneZone,
	LogObjTypeEdgeOneZoneAccelerationDomain,
	LogObjTypeEdgeOneZoneDNSRecord,
	LogObjTypeEdgeOneZoneOriginGroup,
	LogObjTypeTypeEsaSite,
	LogObjTypeTypeEsaSiteOriginPool,
	LogObjTypeTypeEsaSiteEdgeContainerRecord,
	LogObjTypeTypeEsaSiteEdgeRoutineRecord,
	LogObjTypeTypeEsaSiteRecord,
	LogObjTypeIpGroup,
	LogObjTypeIpGroupDevice,
	LogObjTypePlatformConfig,
	LogObjTypeSSLImport,
	LogObjTypeSSLRecord,
	LogObjTypeTUNRecord,
	LogObjTypeWebhook,
	LogObjTypeWebhookSubscription,

	LogObjTypeDeviceIpReport,
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
	LogStatusUnknown string = "unknown" // 未知
	LogStatusSuccess string = "success" // 成功
	LogStatusFailed  string = "failed"  // 失败
)

// GetLogStatusDescriptions 状态说明映射
var GetLogStatusDescriptions = map[string]string{
	LogStatusUnknown: "未知",
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
	DomainMonitorStatusUnknown string = "unknown" // 未知
	DomainMonitorStatusSuccess string = "success" // 成功运行
	DomainMonitorStatusFailure string = "failure" // 检查失败
)

// GetDomainMonitorStatusDescriptions 运行状态说明映射
var GetDomainMonitorStatusDescriptions = map[string]string{
	DomainMonitorStatusUnknown: "未知",
	DomainMonitorStatusSuccess: "成功运行",
	DomainMonitorStatusFailure: "检查失败",
}

// GetDomainMonitorStatuss 返回所有运行状态
var GetDomainMonitorStatuss = []string{
	DomainMonitorStatusUnknown,
	DomainMonitorStatusSuccess,
	DomainMonitorStatusFailure,
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
	SSLRecordStatusUnknown string = "unknown" // 未知
	SSLRecordStatusValid   string = "valid"   // 有效
	SSLRecordStatusExpired string = "expired" // 过期
	SSLRecordStatusRevoked string = "revoked" // 吊销
	SSLRecordStatusPending string = "pending" // 待处理
)

// GetSSLRecordStatusDescriptions SSL记录状态说明映射
var GetSSLRecordStatusDescriptions = map[string]string{
	SSLRecordStatusUnknown: "未知",
	SSLRecordStatusValid:   "有效",
	SSLRecordStatusExpired: "过期",
	SSLRecordStatusRevoked: "吊销",
	SSLRecordStatusPending: "待处理",
}

// GetSSLRecordStatuss 返回所有SSL记录状态
var GetSSLRecordStatuss = []string{
	SSLRecordStatusUnknown,
	SSLRecordStatusValid,
	SSLRecordStatusExpired,
	SSLRecordStatusRevoked,
	SSLRecordStatusPending,
}

const (
	SSLRecordProviderUnknown     string = "unknown"     // 未知
	SSLRecordProviderLetsencrypt string = "letsencrypt" // Letsencrypt 提供商
)

// GetSSLRecordProviderDescriptions SSL记录提供商说明映射
var GetSSLRecordProviderDescriptions = map[string]string{
	SSLRecordProviderUnknown:     "未知",
	SSLRecordProviderLetsencrypt: "Letsencrypt提供商",
}

// GetSSLRecordProviders 返回所有SSL记录提供商
var GetSSLRecordProviders = []string{
	SSLRecordProviderUnknown,
	SSLRecordProviderLetsencrypt,
}

// --------

const (
	SSLImportSourceUnknown string = "unknown" // 未知
	SSLImportSourceWebhook string = "webhook" // Webhook 导入
	SSLImportSourceManual  string = "manual"  // 手动导入
)

// GetSSLImportSourceDescriptions SSL导入来源说明映射
var GetSSLImportSourceDescriptions = map[string]string{
	SSLImportSourceUnknown: "未知",
	SSLImportSourceWebhook: "Webhook导入",
	SSLImportSourceManual:  "手动导入",
}

// GetSSLImportSources 返回所有SSL导入来源
var GetSSLImportSources = []string{
	SSLImportSourceUnknown,
	SSLImportSourceWebhook,
	SSLImportSourceManual,
}

const (
	SSLImportStatusUnknown string = "unknown" // 未知
	SSLImportStatusValid   string = "valid"   // 有效
	SSLImportStatusExpired string = "expired" // 过期
	SSLImportStatusRevoked string = "revoked" // 吊销
	SSLImportStatusPending string = "pending" // 待处理
)

// GetSSLImportStatusDescriptions SSL记录状态说明映射
var GetSSLImportStatusDescriptions = map[string]string{
	SSLImportStatusUnknown: "未知",
	SSLImportStatusValid:   "有效",
	SSLImportStatusExpired: "过期",
	SSLImportStatusRevoked: "吊销",
	SSLImportStatusPending: "待处理",
}

// GetSSLImportStatuss 返回所有SSL记录状态
var GetSSLImportStatuss = []string{
	SSLImportStatusUnknown,
	SSLImportStatusValid,
	SSLImportStatusExpired,
	SSLImportStatusRevoked,
	SSLImportStatusPending,
}

// --------

const (
	TUNRecordServeTypeHTTP        string = "http"        // HTTP
	TUNRecordServeTypeHTTPS       string = "https"       // HTTPS
	TUNRecordServeTypeUNIX        string = "unix"        // UNIX
	TUNRecordServeTypeTCP         string = "tcp"         // TCP
	TUNRecordServeTypeSSH         string = "ssh"         // SSH
	TUNRecordServeTypeRDP         string = "rdp"         // RDP
	TUNRecordServeTypeUNIX_TLS    string = "unix_tls"    // UNIX+TLS
	TUNRecordServeTypeSMB         string = "smb"         // SMB
	TUNRecordServeTypeHTTP_STATUS string = "http_status" // HTTP状态
	TUNRecordServeTypeBASTION     string = "bastion"     // Bastion 主机
)

// GetTUNRecordServeTypeDescriptions 服务类型说明映射
var GetTUNRecordServeTypeDescriptions = map[string]string{
	TUNRecordServeTypeHTTP:        "HTTP",
	TUNRecordServeTypeHTTPS:       "HTTPS",
	TUNRecordServeTypeUNIX:        "UNIX",
	TUNRecordServeTypeTCP:         "TCP",
	TUNRecordServeTypeSSH:         "SSH",
	TUNRecordServeTypeRDP:         "RDP",
	TUNRecordServeTypeUNIX_TLS:    "UNIX+TLS",
	TUNRecordServeTypeSMB:         "SMB",
	TUNRecordServeTypeHTTP_STATUS: "HTTP状态",
	TUNRecordServeTypeBASTION:     "Bastion主机",
}

// GetTUNRecordServeTypes 返回所有服务类型
var GetTUNRecordServeTypes = []string{
	TUNRecordServeTypeHTTP,
	TUNRecordServeTypeHTTPS,
	TUNRecordServeTypeUNIX,
	TUNRecordServeTypeTCP,
	TUNRecordServeTypeSSH,
	TUNRecordServeTypeRDP,
	TUNRecordServeTypeUNIX_TLS,
	TUNRecordServeTypeSMB,
	TUNRecordServeTypeHTTP_STATUS,
	TUNRecordServeTypeBASTION,
}

// --------

const (
	CloudflareTunnelStatusInactive string = "inactive" // 活跃
	CloudflareTunnelStatusDegraded string = "degraded" // 降级
	CloudflareTunnelStatusHealthy  string = "healthy"  // 健康
	CloudflareTunnelStatusDown     string = "down"     // 下线
)

// GetCloudflareTunnelStatusDescriptions 隧道状态 说明映射
var GetCloudflareTunnelStatusDescriptions = map[string]string{
	CloudflareTunnelStatusInactive: "活跃",
	CloudflareTunnelStatusDegraded: "降级",
	CloudflareTunnelStatusHealthy:  "健康",
	CloudflareTunnelStatusDown:     "下线",
}

// GetCloudflareTunnelStatuss 返回所有 隧道状态
var GetCloudflareTunnelStatuss = []string{
	CloudflareTunnelStatusInactive,
	CloudflareTunnelStatusDegraded,
	CloudflareTunnelStatusHealthy,
	CloudflareTunnelStatusDown,
}

const (
	CloudflareTunnelTunTypeCfdTunnel     string = "cfd_tunnel"     // Cloudflare 隧道
	CloudflareTunnelTunTypeWarpConnector string = "warp_connector" // Warp 连接器
	CloudflareTunnelTunTypeWarp          string = "warp"           // Warp 隧道
	CloudflareTunnelTunTypeMagic         string = "magic"          // Magic 隧道
	CloudflareTunnelTunTypeIPSec         string = "ipsec"          // IPsec 隧道
	CloudflareTunnelTunTypeGRE           string = "gre"            // GRE 隧道
	CloudflareTunnelTunTypeCNI           string = "cni"            // CNI 隧道
)

// GetCloudflareTunnelTunTypeDescriptions 隧道类型 说明映射
var GetCloudflareTunnelTunTypeDescriptions = map[string]string{
	CloudflareTunnelTunTypeCfdTunnel:     "Cloudflare 隧道",
	CloudflareTunnelTunTypeWarpConnector: "Warp 连接器",
	CloudflareTunnelTunTypeWarp:          "Warp 隧道",
	CloudflareTunnelTunTypeMagic:         "Magic 隧道",
	CloudflareTunnelTunTypeIPSec:         "IPsec 隧道",
	CloudflareTunnelTunTypeGRE:           "GRE 隧道",
	CloudflareTunnelTunTypeCNI:           "CNI 隧道",
}

// GetCloudflareTunnelTunTypes 返回所有 隧道类型
var GetCloudflareTunnelTunTypes = []string{
	CloudflareTunnelTunTypeCfdTunnel,
	CloudflareTunnelTunTypeWarpConnector,
	CloudflareTunnelTunTypeWarp,
	CloudflareTunnelTunTypeMagic,
	CloudflareTunnelTunTypeIPSec,
	CloudflareTunnelTunTypeGRE,
	CloudflareTunnelTunTypeCNI,
}

// --------

const (
	CloudflareZoneStatusInitializing string = "initializing" // 初始化中
	CloudflareZoneStatusPending      string = "pending"      // 待处理
	CloudflareZoneStatusActive       string = "active"       // 已激活
	CloudflareZoneStatusMoved        string = "moved"        // 已迁移
)

// GetCloudflareZoneStatusDescriptions Cloudflare上的区域状态 说明映射
var GetCloudflareZoneStatusDescriptions = map[string]string{
	CloudflareZoneStatusInitializing: "初始化中",
	CloudflareZoneStatusPending:      "待处理",
	CloudflareZoneStatusActive:       "已激活",
	CloudflareZoneStatusMoved:        "已迁移",
}

// GetCloudflareZoneStatuss 返回所有 Cloudflare上的区域状态
var GetCloudflareZoneStatuss = []string{
	CloudflareZoneStatusInitializing,
	CloudflareZoneStatusPending,
	CloudflareZoneStatusActive,
	CloudflareZoneStatusMoved,
}

const (
	CloudflareZoneTypeFull      string = "full"      // 完整
	CloudflareZoneTypePartial   string = "partial"   // 部分
	CloudflareZoneTypeSecondary string = "secondary" // 二级
	CloudflareZoneTypeInternal  string = "internal"  // 内部
)

// GetCloudflareZoneTypeDescriptions Cloudflare上的区域状态 说明映射
var GetCloudflareZoneTypeDescriptions = map[string]string{
	CloudflareZoneTypeFull:      "完整",
	CloudflareZoneTypePartial:   "部分",
	CloudflareZoneTypeSecondary: "二级",
	CloudflareZoneTypeInternal:  "内部",
}

// GetCloudflareZoneTypes 返回所有 Cloudflare上的区域状态
var GetCloudflareZoneTypes = []string{
	CloudflareZoneTypeFull,
	CloudflareZoneTypePartial,
	CloudflareZoneTypeSecondary,
	CloudflareZoneTypeInternal,
}

const (
	CloudflareZoneRecordTypeA      string = "A"      // A
	CloudflareZoneRecordTypeAaaa   string = "AAAA"   // AAAA
	CloudflareZoneRecordTypeCname  string = "CNAME"  // CNAME
	CloudflareZoneRecordTypeMx     string = "MX"     // MX
	CloudflareZoneRecordTypeNs     string = "NS"     // NS
	CloudflareZoneRecordTypePtr    string = "PTR"    // PTR
	CloudflareZoneRecordTypeTxt    string = "TXT"    // TXT
	CloudflareZoneRecordTypeCaa    string = "CAA"    // CAA
	CloudflareZoneRecordTypeCert   string = "CERT"   // CERT
	CloudflareZoneRecordTypeDnskey string = "DNSKEY" // DNSKEY
	CloudflareZoneRecordTypeDs     string = "DS"     // DS
	CloudflareZoneRecordTypeHttps  string = "HTTPS"  // HTTPS
	CloudflareZoneRecordTypeLoc    string = "LOC"    // LOC
	CloudflareZoneRecordTypeNaptr  string = "NAPTR"  // NAPTR
	CloudflareZoneRecordTypeSmimea string = "SMIMEA" // SMIMEA
	CloudflareZoneRecordTypeSrv    string = "SRV"    // SRV
	CloudflareZoneRecordTypeSshfp  string = "SSHFP"  // SSHFP
	CloudflareZoneRecordTypeSvc    string = "SVCB"   // SVCB
	CloudflareZoneRecordTypeTlsa   string = "TLSA"   // TLSA
	CloudflareZoneRecordTypeUri    string = "URI"    // URI
)

// GetCloudflareZoneRecordTypeDescriptions 数据录类型 说明映射
var GetCloudflareZoneRecordTypeDescriptions = map[string]string{
	CloudflareZoneRecordTypeA:      "A",
	CloudflareZoneRecordTypeAaaa:   "AAAA",
	CloudflareZoneRecordTypeCname:  "CNAME",
	CloudflareZoneRecordTypeMx:     "MX",
	CloudflareZoneRecordTypeNs:     "NS",
	CloudflareZoneRecordTypePtr:    "PTR",
	CloudflareZoneRecordTypeTxt:    "TXT",
	CloudflareZoneRecordTypeCaa:    "CAA",
	CloudflareZoneRecordTypeCert:   "CERT",
	CloudflareZoneRecordTypeDnskey: "DNSKEY",
	CloudflareZoneRecordTypeDs:     "DS",
	CloudflareZoneRecordTypeHttps:  "HTTPS",
	CloudflareZoneRecordTypeLoc:    "LOC",
	CloudflareZoneRecordTypeNaptr:  "NAPTR",
	CloudflareZoneRecordTypeSmimea: "SMIMEA",
	CloudflareZoneRecordTypeSrv:    "SRV",
	CloudflareZoneRecordTypeSshfp:  "SSHFP",
	CloudflareZoneRecordTypeSvc:    "SVCB",
	CloudflareZoneRecordTypeTlsa:   "TLSA",
	CloudflareZoneRecordTypeUri:    "URI",
}

// GetCloudflareZoneRecordTypes 返回所有 数据录类型
var GetCloudflareZoneRecordTypes = []string{
	CloudflareZoneRecordTypeA,
	CloudflareZoneRecordTypeAaaa,
	CloudflareZoneRecordTypeCname,
	CloudflareZoneRecordTypeMx,
	CloudflareZoneRecordTypeNs,
	CloudflareZoneRecordTypePtr,
	CloudflareZoneRecordTypeTxt,
	CloudflareZoneRecordTypeCaa,
	CloudflareZoneRecordTypeCert,
	CloudflareZoneRecordTypeDnskey,
	CloudflareZoneRecordTypeDs,
	CloudflareZoneRecordTypeHttps,
	CloudflareZoneRecordTypeLoc,
	CloudflareZoneRecordTypeNaptr,
	CloudflareZoneRecordTypeSmimea,
	CloudflareZoneRecordTypeSrv,
	CloudflareZoneRecordTypeSshfp,
	CloudflareZoneRecordTypeSvc,
	CloudflareZoneRecordTypeTlsa,
	CloudflareZoneRecordTypeUri,
}

// --------

const (
	EdgeOneZoneAreaGlobal   string = "global"   // 全球可用区
	EdgeOneZoneAreaMainland string = "mainland" // 中国大陆可用区
	EdgeOneZoneAreaOverseas string = "overseas" // 全球可用区（不含中国大陆）
)

// GetEdgeOneZoneAreaDescriptions 加速区域 说明映射
var GetEdgeOneZoneAreaDescriptions = map[string]string{
	EdgeOneZoneAreaGlobal:   "全球可用区",
	EdgeOneZoneAreaMainland: "中国大陆可用区",
	EdgeOneZoneAreaOverseas: "全球可用区（不含中国大陆）",
}

// GetEdgeOneZoneAreas 返回所有 加速区域
var GetEdgeOneZoneAreas = []string{
	EdgeOneZoneAreaGlobal,
	EdgeOneZoneAreaMainland,
	EdgeOneZoneAreaOverseas,
}

const (
	EdgeOneZoneTypeFull           string = "full"           // NS 接入
	EdgeOneZoneTypePartial        string = "partial"        // CNAME 接入
	EdgeOneZoneTypeNoDomainAccess string = "noDomainAccess" // 无域名接入
	EdgeOneZoneTypeDnsPodAccess   string = "dnsPodAccess"   // DNSPod 托管
	EdgeOneZoneTypePages          string = "pages"          // Pages
)

// GetEdgeOneZoneTypeDescriptions 接入类型 说明映射
var GetEdgeOneZoneTypeDescriptions = map[string]string{
	EdgeOneZoneTypeFull:           "NS 接入",
	EdgeOneZoneTypePartial:        "CNAME 接入",
	EdgeOneZoneTypeNoDomainAccess: "无域名接入",
	EdgeOneZoneTypeDnsPodAccess:   "DNSPod 托管",
	EdgeOneZoneTypePages:          "Pages",
}

// GetEdgeOneZoneTypes 返回所有 接入类型
var GetEdgeOneZoneTypes = []string{
	EdgeOneZoneTypeFull,
	EdgeOneZoneTypePartial,
	EdgeOneZoneTypeNoDomainAccess,
	EdgeOneZoneTypeDnsPodAccess,
	EdgeOneZoneTypePages,
}

const (
	EdgeOneZoneStatusActive       string = "active"       // NS 已切换
	EdgeOneZoneStatusPending      string = "pending"      // NS 未切换
	EdgeOneZoneStatusMoved        string = "moved"        // NS 已切走
	EdgeOneZoneStatusDeactivated  string = "deactivated"  // 被封禁
	EdgeOneZoneStatusInitializing string = "initializing" // 待绑定套餐
)

// GetEdgeOneZoneStatusDescriptions 站点状态 说明映射
var GetEdgeOneZoneStatusDescriptions = map[string]string{
	EdgeOneZoneStatusActive:       "NS 已切换",
	EdgeOneZoneStatusPending:      "NS 未切换",
	EdgeOneZoneStatusMoved:        "NS 已切走",
	EdgeOneZoneStatusDeactivated:  "被封禁",
	EdgeOneZoneStatusInitializing: "待绑定套餐",
}

// GetEdgeOneZoneStatuss 返回所有 站点状态
var GetEdgeOneZoneStatuss = []string{
	EdgeOneZoneStatusActive,
	EdgeOneZoneStatusPending,
	EdgeOneZoneStatusMoved,
	EdgeOneZoneStatusDeactivated,
	EdgeOneZoneStatusInitializing,
}

const (
	EdgeOneZoneCnameStatusFinished string = "finished" // 站点已验证
	EdgeOneZoneCnameStatusPending  string = "pending"  // 站点验证中
)

// GetEdgeOneZoneCnameStatusDescriptions CNAME 接入状态 说明映射
var GetEdgeOneZoneCnameStatusDescriptions = map[string]string{
	EdgeOneZoneCnameStatusFinished: "站点已验证",
	EdgeOneZoneCnameStatusPending:  "站点验证中",
}

// GetEdgeOneZoneCnameStatuss 返回所有 CNAME 接入状态
var GetEdgeOneZoneCnameStatuss = []string{
	EdgeOneZoneCnameStatusFinished,
	EdgeOneZoneCnameStatusPending,
}

const (
	EdgeOneZoneActiveStatusActive   string = "active"   // 已启用
	EdgeOneZoneActiveStatusInactive string = "inactive" // 未生效
	EdgeOneZoneActiveStatusPaused   string = "paused"   // 已停用
)

// GetEdgeOneZoneActiveStatusDescriptions 展示状态 说明映射
var GetEdgeOneZoneActiveStatusDescriptions = map[string]string{
	EdgeOneZoneActiveStatusActive:   "已启用",
	EdgeOneZoneActiveStatusInactive: "未生效",
	EdgeOneZoneActiveStatusPaused:   "已停用",
}

// GetEdgeOneZoneActiveStatuss 返回所有 展示状态
var GetEdgeOneZoneActiveStatuss = []string{
	EdgeOneZoneActiveStatusActive,
	EdgeOneZoneActiveStatusInactive,
	EdgeOneZoneActiveStatusPaused,
}

const (
	EdgeOneZoneLockStatusEnable      string = "enable"       // 正常
	EdgeOneZoneLockStatusDisable     string = "disable"      // 锁定中
	EdgeOneZoneLockStatusPlanMigrate string = "plan_migrate" // 套餐迁移中
)

// GetEdgeOneZoneLockStatusDescriptions 锁定状态 说明映射
var GetEdgeOneZoneLockStatusDescriptions = map[string]string{
	EdgeOneZoneLockStatusEnable:      "正常",
	EdgeOneZoneLockStatusDisable:     "锁定中",
	EdgeOneZoneLockStatusPlanMigrate: "套餐迁移中",
}

// GetEdgeOneZoneLockStatuss 返回所有 锁定状态
var GetEdgeOneZoneLockStatuss = []string{
	EdgeOneZoneLockStatusEnable,
	EdgeOneZoneLockStatusDisable,
	EdgeOneZoneLockStatusPlanMigrate,
}

// --------

const (
	EdgeOneZoneAccelerationDomainDomainStatusOnline    string = "online"    // 已生效
	EdgeOneZoneAccelerationDomainDomainStatusProcess   string = "process"   // 部署中
	EdgeOneZoneAccelerationDomainDomainStatusOffline   string = "offline"   // 已停用
	EdgeOneZoneAccelerationDomainDomainStatusForbidden string = "forbidden" // 已封禁
	EdgeOneZoneAccelerationDomainDomainStatusInit      string = "init"      // 未生效，待激活站点
)

// GetEdgeOneZoneAccelerationDomainDomainStatusDescriptions 加速域名状态 说明映射
var GetEdgeOneZoneAccelerationDomainDomainStatusDescriptions = map[string]string{
	EdgeOneZoneAccelerationDomainDomainStatusOnline:    "已生效",
	EdgeOneZoneAccelerationDomainDomainStatusProcess:   "部署中",
	EdgeOneZoneAccelerationDomainDomainStatusOffline:   "已停用",
	EdgeOneZoneAccelerationDomainDomainStatusForbidden: "已封禁",
	EdgeOneZoneAccelerationDomainDomainStatusInit:      "未生效，待激活站点",
}

// GetEdgeOneZoneAccelerationDomainDomainStatuss 返回所有 加速域名状态
var GetEdgeOneZoneAccelerationDomainDomainStatuss = []string{
	EdgeOneZoneAccelerationDomainDomainStatusOnline,
	EdgeOneZoneAccelerationDomainDomainStatusProcess,
	EdgeOneZoneAccelerationDomainDomainStatusOffline,
	EdgeOneZoneAccelerationDomainDomainStatusForbidden,
	EdgeOneZoneAccelerationDomainDomainStatusInit,
}

const (
	EdgeOneZoneAccelerationDomainIPv6StatusFollow string = "follow" // 遵循站点
	EdgeOneZoneAccelerationDomainIPv6StatusOn     string = "on"     // 开启
	EdgeOneZoneAccelerationDomainIPv6StatusOff    string = "off"    // 关闭
)

// GetEdgeOneZoneAccelerationDomainIPv6StatusDescriptions IPv6 状态 说明映射
var GetEdgeOneZoneAccelerationDomainIPv6StatusDescriptions = map[string]string{
	EdgeOneZoneAccelerationDomainIPv6StatusFollow: "遵循站点",
	EdgeOneZoneAccelerationDomainIPv6StatusOn:     "开启",
	EdgeOneZoneAccelerationDomainIPv6StatusOff:    "关闭",
}

// GetEdgeOneZoneAccelerationDomainIPv6Statuss 返回所有 IPv6 状态
var GetEdgeOneZoneAccelerationDomainIPv6Statuss = []string{
	EdgeOneZoneAccelerationDomainIPv6StatusFollow,
	EdgeOneZoneAccelerationDomainIPv6StatusOn,
	EdgeOneZoneAccelerationDomainIPv6StatusOff,
}

const (
	EdgeOneZoneAccelerationDomainIdentificationStatusPending  string = "pending"  // 待验证
	EdgeOneZoneAccelerationDomainIdentificationStatusFinished string = "finished" // 已完成验证
)

// GetEdgeOneZoneAccelerationDomainIdentificationStatusDescriptions 加速域名归属权验证状态 说明映射
var GetEdgeOneZoneAccelerationDomainIdentificationStatusDescriptions = map[string]string{
	EdgeOneZoneAccelerationDomainIdentificationStatusPending:  "待验证",
	EdgeOneZoneAccelerationDomainIdentificationStatusFinished: "已完成验证",
}

// GetEdgeOneZoneAccelerationDomainIdentificationStatuss 返回所有 加速域名归属权验证状态
var GetEdgeOneZoneAccelerationDomainIdentificationStatuss = []string{
	EdgeOneZoneAccelerationDomainIdentificationStatusPending,
	EdgeOneZoneAccelerationDomainIdentificationStatusFinished,
}

const (
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeIpDomain    string = "IP_DOMAIN"    // IPV4、IPV6 或域名类型源站
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeCos         string = "COS"          // 腾讯云 COS 对象存储源站
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeAwsS3       string = "AWS_S3"       // AWS S3 对象存储源站
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeOriginGroup string = "ORIGIN_GROUP" // 源站组类型源站
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeVod         string = "VOD"          // 云点播
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeSpace       string = "SPACE"        // 源站卸载
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeLb          string = "LB"           // 负载均衡
)

// GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypeDescriptions 源站信息 源站类型 说明映射
var GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypeDescriptions = map[string]string{
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeIpDomain:    "IP/域名",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeCos:         "对象存储",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeAwsS3:       "对象存储",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeOriginGroup: "源站组",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeVod:         "云点播",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeSpace:       "源站卸载",
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeLb:          "负载均衡",
}

// GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypes 返回所有 源站信息 源站类型
var GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypes = []string{
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeIpDomain,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeCos,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeAwsS3,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeOriginGroup,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeVod,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeSpace,
	EdgeOneZoneAccelerationDomainOriginDetailOriginTypeLb,
}

const (
	EdgeOneZoneAccelerationDomainOriginProtocolFollow string = "FOLLOW" // 协议跟随
	EdgeOneZoneAccelerationDomainOriginProtocolHttp   string = "HTTP"   // HTTP协议回源
	EdgeOneZoneAccelerationDomainOriginProtocolHttps  string = "HTTPS"  // HTTPS协议回源
)

// GetEdgeOneZoneAccelerationDomainOriginProtocolDescriptions 回源协议 说明映射
var GetEdgeOneZoneAccelerationDomainOriginProtocolDescriptions = map[string]string{
	EdgeOneZoneAccelerationDomainOriginProtocolFollow: "协议跟随",
	EdgeOneZoneAccelerationDomainOriginProtocolHttp:   "HTTP协议回源",
	EdgeOneZoneAccelerationDomainOriginProtocolHttps:  "HTTPS协议回源",
}

// GetEdgeOneZoneAccelerationDomainOriginProtocols 返回所有 回源协议
var GetEdgeOneZoneAccelerationDomainOriginProtocols = []string{
	EdgeOneZoneAccelerationDomainOriginProtocolFollow,
	EdgeOneZoneAccelerationDomainOriginProtocolHttp,
	EdgeOneZoneAccelerationDomainOriginProtocolHttps,
}

// --------

const (
	EdgeOneZoneDNSRecordTypeA     string = "A"     // A
	EdgeOneZoneDNSRecordTypeAAAA  string = "AAAA"  // AAAA
	EdgeOneZoneDNSRecordTypeMX    string = "MX"    // MX
	EdgeOneZoneDNSRecordTypeCNAME string = "CNAME" // CNAME
	EdgeOneZoneDNSRecordTypeTXT   string = "TXT"   // TXT
	EdgeOneZoneDNSRecordTypeNS    string = "NS"    // NS
	EdgeOneZoneDNSRecordTypeCAA   string = "CAA"   // CAA
	EdgeOneZoneDNSRecordTypeSRV   string = "SRV"   // SRV
)

// GetEdgeOneZoneDNSRecordTypeDescriptions DNS 记录类型 说明映射
var GetEdgeOneZoneDNSRecordTypeDescriptions = map[string]string{
	EdgeOneZoneDNSRecordTypeA:     "A",
	EdgeOneZoneDNSRecordTypeAAAA:  "AAAA",
	EdgeOneZoneDNSRecordTypeMX:    "MX",
	EdgeOneZoneDNSRecordTypeCNAME: "CNAME",
	EdgeOneZoneDNSRecordTypeTXT:   "TXT",
	EdgeOneZoneDNSRecordTypeNS:    "NS",
	EdgeOneZoneDNSRecordTypeCAA:   "CAA",
	EdgeOneZoneDNSRecordTypeSRV:   "SRV",
}

// GetEdgeOneZoneDNSRecordTypes 返回所有 DNS 记录类型
var GetEdgeOneZoneDNSRecordTypes = []string{
	EdgeOneZoneDNSRecordTypeA,
	EdgeOneZoneDNSRecordTypeAAAA,
	EdgeOneZoneDNSRecordTypeMX,
	EdgeOneZoneDNSRecordTypeCNAME,
	EdgeOneZoneDNSRecordTypeTXT,
	EdgeOneZoneDNSRecordTypeNS,
	EdgeOneZoneDNSRecordTypeCAA,
	EdgeOneZoneDNSRecordTypeSRV,
}

const (
	EdgeOneZoneDNSRecordStatusEnable  string = "enable"  // 已生效
	EdgeOneZoneDNSRecordStatusDisable string = "disable" // 已停用
)

// GetEdgeOneZoneDNSRecordStatusDescriptions DNS 记录类型 说明映射
var GetEdgeOneZoneDNSRecordStatusDescriptions = map[string]string{
	EdgeOneZoneDNSRecordStatusEnable:  "已生效",
	EdgeOneZoneDNSRecordStatusDisable: "已停用",
}

// GetEdgeOneZoneDNSRecordStatuss 返回所有 DNS 记录类型
var GetEdgeOneZoneDNSRecordStatuss = []string{
	EdgeOneZoneDNSRecordStatusEnable,
	EdgeOneZoneDNSRecordStatusDisable,
}

// --------

const (
	EdgeOneZoneOriginGroupTypeGeneral string = "GENERAL" // 通用型
	EdgeOneZoneOriginGroupTypeHttp    string = "HTTP"    //  HTTP专用型
)

// GetEdgeOneZoneOriginGroupTypeDescriptions 源站组类型 说明映射
var GetEdgeOneZoneOriginGroupTypeDescriptions = map[string]string{
	EdgeOneZoneOriginGroupTypeGeneral: "通用型",
	EdgeOneZoneOriginGroupTypeHttp:    "HTTP专用型",
}

// GetEdgeOneZoneOriginGroupTypes 返回所有 源站组类型
var GetEdgeOneZoneOriginGroupTypes = []string{
	EdgeOneZoneOriginGroupTypeGeneral,
	EdgeOneZoneOriginGroupTypeHttp,
}

// --------

const (
	EsaSiteAccessTypeNs    string = "NS"    // NS
	EsaSiteAccessTypeCname string = "CNAME" // CNAME
)

// GetEsaSiteAccessTypeDescriptions 站点接入类型 说明映射
var GetEsaSiteAccessTypeDescriptions = map[string]string{
	EsaSiteAccessTypeNs:    "NS",
	EsaSiteAccessTypeCname: "CNAME",
}

// GetEsaSiteAccessTypes 返回所有 站点接入类型
var GetEsaSiteAccessTypes = []string{
	EsaSiteAccessTypeNs,
	EsaSiteAccessTypeCname,
}

const (
	EsaSiteCoverageDomestic string = "domestic" // 仅中国内地
	EsaSiteCoverageGlobal   string = "global"   // 全球
	EsaSiteCoverageOverseas string = "overseas" // 全球（不包含中国内地）
)

// GetEsaSiteCoverageDescriptions 站点加速区域 说明映射
var GetEsaSiteCoverageDescriptions = map[string]string{
	EsaSiteCoverageDomestic: "仅中国内地",
	EsaSiteCoverageGlobal:   "全球",
	EsaSiteCoverageOverseas: "全球（不包含中国内地）",
}

// GetEsaSiteCoverages 返回所有 站点加速区域
var GetEsaSiteCoverages = []string{
	EsaSiteCoverageDomestic,
	EsaSiteCoverageGlobal,
	EsaSiteCoverageOverseas,
}

const (
	EsaSiteStatusPending string = "pending" // 站点待配置
	EsaSiteStatusActive  string = "active"  // 站点已激活
	EsaSiteStatusOffline string = "offline" // 站点已下线
	EsaSiteStatusMoved   string = "moved"   // 站点已被顶替
)

// GetEsaSiteStatusDescriptions 站点状态 说明映射
var GetEsaSiteStatusDescriptions = map[string]string{
	EsaSiteStatusPending: "站点待配置",
	EsaSiteStatusActive:  "站点已激活",
	EsaSiteStatusOffline: "站点已下线",
	EsaSiteStatusMoved:   "站点已被顶替",
}

// GetEsaSiteStatuss 返回所有 站点状态
var GetEsaSiteStatuss = []string{
	EsaSiteStatusPending,
	EsaSiteStatusActive,
	EsaSiteStatusOffline,
	EsaSiteStatusMoved,
}

// --------

const (
	EsaSiteRecordBizNameImageVideo string = "image_video" // 视频图片
	EsaSiteRecordBizNameApi        string = "api"         // api 接口
	EsaSiteRecordBizNameWeb        string = "web"         // web 网页
)

// GetEsaSiteRecordBizNameDescriptions 记录加速时的业务场景 说明映射
var GetEsaSiteRecordBizNameDescriptions = map[string]string{
	EsaSiteRecordBizNameImageVideo: "视频图片",
	EsaSiteRecordBizNameApi:        "api 接口",
	EsaSiteRecordBizNameWeb:        "web 网页",
}

// GetEsaSiteRecordBizNames 返回所有 记录加速时的业务场景
var GetEsaSiteRecordBizNames = []string{
	EsaSiteRecordBizNameImageVideo,
	EsaSiteRecordBizNameApi,
	EsaSiteRecordBizNameWeb,
}

const (
	EsaSiteRecordRecordSourceTypeOss    string = "OSS"    // OSS 源站
	EsaSiteRecordRecordSourceTypeS3     string = "S3"     // S3 源站
	EsaSiteRecordRecordSourceTypeLb     string = "LB"     // 负载均衡器源站
	EsaSiteRecordRecordSourceTypeOp     string = "OP"     // 源地址池源站
	EsaSiteRecordRecordSourceTypeDomain string = "Domain" // 普通域名源站
)

// GetEsaSiteRecordRecordSourceTypeDescriptions CNAME 记录的源站类型 说明映射
var GetEsaSiteRecordRecordSourceTypeDescriptions = map[string]string{
	EsaSiteRecordRecordSourceTypeOss:    "OSS 源站",
	EsaSiteRecordRecordSourceTypeS3:     "S3 源站",
	EsaSiteRecordRecordSourceTypeLb:     "负载均衡器源站",
	EsaSiteRecordRecordSourceTypeOp:     "源地址池源站",
	EsaSiteRecordRecordSourceTypeDomain: "普通域名源站",
}

// GetEsaSiteRecordRecordSourceTypes 返回所有 CNAME 记录的源站类型
var GetEsaSiteRecordRecordSourceTypes = []string{
	EsaSiteRecordRecordSourceTypeOss,
	EsaSiteRecordRecordSourceTypeS3,
	EsaSiteRecordRecordSourceTypeLb,
	EsaSiteRecordRecordSourceTypeOp,
	EsaSiteRecordRecordSourceTypeDomain,
}

const (
	EsaSiteRecordRecordTypeAaaaa string = "A_AAAA" // A/AAAA
	EsaSiteRecordRecordTypeCname string = "CNAME"  // CNAME
	EsaSiteRecordRecordTypeTxt   string = "TXT"    // TXT
)

// GetEsaSiteRecordRecordTypeDescriptions 记录的 DNS 类型 说明映射
var GetEsaSiteRecordRecordTypeDescriptions = map[string]string{
	EsaSiteRecordRecordTypeAaaaa: "AAAAA",
	EsaSiteRecordRecordTypeCname: "CNAME",
	EsaSiteRecordRecordTypeTxt:   "TXT",
}

// GetEsaSiteRecordRecordTypes 返回所有 记录的 DNS 类型
var GetEsaSiteRecordRecordTypes = []string{
	EsaSiteRecordRecordTypeAaaaa,
	EsaSiteRecordRecordTypeCname,
	EsaSiteRecordRecordTypeTxt,
}

const (
	EsaSiteRecordHostPolicyFollowHostname     string = "follow_hostname"      // 跟随请求 HOST
	EsaSiteRecordHostPolicyFollowOriginDomain string = "follow_origin_domain" // 跟随源站域名
)

// GetEsaSiteRecordHostPolicyDescriptions 回源 HOST 策略 说明映射
var GetEsaSiteRecordHostPolicyDescriptions = map[string]string{
	EsaSiteRecordHostPolicyFollowHostname:     "跟随请求 HOST",
	EsaSiteRecordHostPolicyFollowOriginDomain: "跟随源站域名",
}

// GetEsaSiteRecordHostPolicys 返回所有 回源 HOST 策略
var GetEsaSiteRecordHostPolicys = []string{
	EsaSiteRecordHostPolicyFollowHostname,
	EsaSiteRecordHostPolicyFollowOriginDomain,
}

// --------

const (
	CronJobTaskTypeOneOff    string = "one_off"   // 一次性
	CronJobTaskTypeRecurring string = "recurring" // 周期性
)

// GetCronJobTaskTypeDescriptions 任务类型 说明映射
var GetCronJobTaskTypeDescriptions = map[string]string{
	CronJobTaskTypeOneOff:    "一次性",
	CronJobTaskTypeRecurring: "周期性",
}

// GetCronJobTaskTypes 返回所有 任务类型
var GetCronJobTaskTypes = []string{
	CronJobTaskTypeOneOff,
	CronJobTaskTypeRecurring,
}

const (
	CronJobJobTypeCron     string = "cron"     // cron 表达式
	CronJobJobTypeDuration string = "duration" // 固定间隔
	CronJobJobTypeRandom   string = "random"   // 随机间隔
)

// GetCronJobJobTypeDescriptions 任务类型 说明映射
var GetCronJobJobTypeDescriptions = map[string]string{
	CronJobJobTypeCron:     "cron 表达式",
	CronJobJobTypeDuration: "固定间隔",
	CronJobJobTypeRandom:   "随机间隔",
}

// GetCronJobJobTypes 返回所有 任务类型
var GetCronJobJobTypes = []string{
	CronJobJobTypeCron,
	CronJobJobTypeDuration,
	CronJobJobTypeRandom,
}

const (
	CronJobJobCategorySystem string = "system" // 系统内置
	CronJobJobCategoryCustom string = "custom" // 用户自定义
)

// GetCronJobJobCategoryDescriptions 任务分类 说明映射
var GetCronJobJobCategoryDescriptions = map[string]string{
	CronJobJobCategorySystem: "系统内置",
	CronJobJobCategoryCustom: "用户自定义",
}

// GetCronJobJobCategorys 返回所有 任务分类
var GetCronJobJobCategorys = []string{
	CronJobJobCategorySystem,
	CronJobJobCategoryCustom,
}

const (
	CronJobBusinessTypeCdn              string = "cdn"               // cdn
	CronJobBusinessTypeDns              string = "dns"               // dns
	CronJobBusinessTypeSsl              string = "ssl"               // ssl
	CronJobBusinessTypeDomain           string = "domain"            // domain
	CronJobBusinessTypeTun              string = "tun"               // tun
	CronJobBusinessTypeWebhook          string = "webhook"           // webhook
	CronJobBusinessTypeDevice           string = "device"            // device
	CronJobBusinessTypeCloudflareTunnel string = "cloudflare_tunnel" // cloudflare_tunnel
	CronJobBusinessTypeCloudflareZone   string = "cloudflare_zone"   // cloudflare_zone
	CronJobBusinessTypeEdgeOneZone      string = "edgeone_zone"      // edgeone_zone
	CronJobBusinessTypeESAsite          string = "esa_site"          // esa_site
	CronJobBusinessTypeSystem           string = "system"            // system
)

// GetCronJobBusinessTypeDescriptions 业务类型 说明映射
var GetCronJobBusinessTypeDescriptions = map[string]string{
	CronJobBusinessTypeCdn:              "CDN",
	CronJobBusinessTypeDns:              "DNS",
	CronJobBusinessTypeSsl:              "SSL",
	CronJobBusinessTypeDomain:           "域名",
	CronJobBusinessTypeTun:              "TUN",
	CronJobBusinessTypeWebhook:          "通知",
	CronJobBusinessTypeDevice:           "设备",
	CronJobBusinessTypeCloudflareTunnel: "Cloudflare隧道",
	CronJobBusinessTypeCloudflareZone:   "Cloudflare站点",
	CronJobBusinessTypeEdgeOneZone:      "EdgeOne站点",
	CronJobBusinessTypeESAsite:          "ESA站点",
	CronJobBusinessTypeSystem:           "系统",
}

// GetCronJobBusinessTypes 返回所有 业务类型
var GetCronJobBusinessTypes = []string{
	CronJobBusinessTypeCdn,
	CronJobBusinessTypeDns,
	CronJobBusinessTypeSsl,
	CronJobBusinessTypeDomain,
	CronJobBusinessTypeTun,
	CronJobBusinessTypeWebhook,
	CronJobBusinessTypeDevice,
	CronJobBusinessTypeCloudflareTunnel,
	CronJobBusinessTypeCloudflareZone,
	CronJobBusinessTypeEdgeOneZone,
	CronJobBusinessTypeESAsite,
	CronJobBusinessTypeSystem,
}

const (
	CronJobIntervalUnitSeconds string = "seconds" // 秒
	CronJobIntervalUnitMinutes string = "minutes" // 分钟
	CronJobIntervalUnitHours   string = "hours"   // 小时
	CronJobIntervalUnitDays    string = "days"    // 天
	CronJobIntervalUnitWeeks   string = "weeks"   // 周
)

// GetCronJobIntervalUnitDescriptions 间隔时间单位 说明映射
var GetCronJobIntervalUnitDescriptions = map[string]string{
	CronJobIntervalUnitSeconds: "秒",
	CronJobIntervalUnitMinutes: "分钟",
	CronJobIntervalUnitHours:   "小时",
	CronJobIntervalUnitDays:    "天",
	CronJobIntervalUnitWeeks:   "周",
}

// GetCronJobIntervalUnits 返回所有 间隔时间单位
var GetCronJobIntervalUnits = []string{
	CronJobIntervalUnitSeconds,
	CronJobIntervalUnitMinutes,
	CronJobIntervalUnitHours,
	CronJobIntervalUnitDays,
	CronJobIntervalUnitWeeks,
}

const (
	CronJobStatusRunning  string = "running"  // 运行中
	CronJobStatusStopped  string = "stopped"  // 已停止
	CronJobStatusFinished string = "finished" // 已完成
	CronJobStatusFailed   string = "failed"   // 已失败
)

// GetCronJobStatusDescriptions 状态 说明映射
var GetCronJobStatusDescriptions = map[string]string{
	CronJobStatusRunning:  "运行中",
	CronJobStatusStopped:  "已停止",
	CronJobStatusFinished: "已完成",
	CronJobStatusFailed:   "已失败",
}

// GetCronJobStatuss 返回所有 状态
var GetCronJobStatuss = []string{
	CronJobStatusRunning,
	CronJobStatusStopped,
	CronJobStatusFinished,
	CronJobStatusFailed,
}

// --------

const (
	CronJobLogStatusSuccess string = "success" // 成功
	CronJobLogStatusFailed  string = "failed"  // 失败
	CronJobLogStatusRunning string = "running" // 运行
)

// GetCronJobLogStatusDescriptions 状态 说明映射
var GetCronJobLogStatusDescriptions = map[string]string{
	CronJobLogStatusSuccess: "成功",
	CronJobLogStatusFailed:  "失败",
	CronJobLogStatusRunning: "运行",
}

// GetCronJobLogStatuss 返回所有 状态
var GetCronJobLogStatuss = []string{
	CronJobLogStatusSuccess,
	CronJobLogStatusFailed,
	CronJobLogStatusRunning,
}
