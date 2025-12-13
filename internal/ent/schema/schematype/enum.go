package schematype

const (
	OriginProtocolHTTP   string = "http"   // HTTP协议
	OriginProtocolHTTPS  string = "https"  // HTTPS协议
	OriginProtocolFollow string = "follow" // 跟随协议
)

// GetOriginProtocolDescriptions 回源协议说明映射
var GetOriginProtocolDescriptions = map[string]string{
	OriginProtocolHTTP:   "HTTP协议",
	OriginProtocolHTTPS:  "HTTPS协议",
	OriginProtocolFollow: "跟随协议",
}

// GetOriginProtocols 返回回源协议
var GetOriginProtocols = []string{
	OriginProtocolHTTP,
	OriginProtocolHTTPS,
	OriginProtocolFollow,
}

// --------

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
	ConfigCapabilityDns string = "dns" // DNS服务能力
	ConfigCapabilityCdn string = "cdn" // CDN服务能力
	ConfigCapabilityAll string = "all" // 所有服务能力
)

// GetConfigCapabilityDescriptions 服务能力说明映射
var GetConfigCapabilityDescriptions = map[string]string{
	ConfigCapabilityDns: "DNS服务能力",
	ConfigCapabilityCdn: "CDN服务能力",
	ConfigCapabilityAll: "所有服务能力",
}

// GetConfigCapabilitys 返回所有服务能力
var GetConfigCapabilitys = []string{
	ConfigCapabilityDns,
	ConfigCapabilityCdn,
	ConfigCapabilityAll,
}

// --------

const (
	ConfigPlatformCloudflare string = "cloudflare" // Cloudflare
	ConfigPlatformEdgeOne    string = "edgeone"    // EdgeOne
	ConfigPlatformDnspod     string = "dnspod"     // DNSPod
)

// GetConfigPlatformDescriptions 平台类型说明映射
var GetConfigPlatformDescriptions = map[string]string{
	ConfigPlatformCloudflare: "Cloudflare",
	ConfigPlatformEdgeOne:    "EdgeOne",
	ConfigPlatformDnspod:     "DNSPod",
}

// GetConfigPlatforms 返回所有平台类型
var GetConfigPlatforms = []string{
	ConfigPlatformCloudflare,
	ConfigPlatformEdgeOne,
	ConfigPlatformDnspod,
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
	WebhookSubscriptionObjTypeCdnRecord     string = "cdn_record"     // CDN记录
	WebhookSubscriptionObjTypeDnsRecord     string = "dns_record"     // DNS记录
	WebhookSubscriptionObjTypeDevice        string = "device"         // 设备
	WebhookSubscriptionObjTypeIPGroup       string = "ip_group"       // IP组
	WebhookSubscriptionObjTypeDomainMonitor string = "domain_monitor" // 域名监控
)

// GetWebhookSubscriptionObjTypeDescriptions Webhook 订阅对象类型说明映射
var GetWebhookSubscriptionObjTypeDescriptions = map[string]string{
	WebhookSubscriptionObjTypeCdnRecord:     "CDN记录",
	WebhookSubscriptionObjTypeDnsRecord:     "DNS记录",
	WebhookSubscriptionObjTypeDevice:        "设备",
	WebhookSubscriptionObjTypeIPGroup:       "IP组",
	WebhookSubscriptionObjTypeDomainMonitor: "域名监控",
}

// GetWebhookSubscriptionObjTypes 返回所有 Webhook 订阅对象类型
var GetWebhookSubscriptionObjTypes = []string{
	WebhookSubscriptionObjTypeCdnRecord,
	WebhookSubscriptionObjTypeDnsRecord,
	WebhookSubscriptionObjTypeDevice,
	WebhookSubscriptionObjTypeIPGroup,
	WebhookSubscriptionObjTypeDomainMonitor,
}

// --------

const (
	LogOpTypeDataCreate   string = "data_create"   // 数据创建
	LogOpTypeDataUpdate   string = "data_update"   // 数据更新
	LogOpTypeDataDelete   string = "data_delete"   // 数据删除
	LogOpTypeDeviceReport string = "device_report" // 设备上报
	LogOpTypeIPRefresh    string = "ip_refresh"    // IP刷新
)

// GetLogOpTypeDescriptions 操作类型说明映射
var GetLogOpTypeDescriptions = map[string]string{
	LogOpTypeDataCreate:   "数据创建",
	LogOpTypeDataUpdate:   "数据更新",
	LogOpTypeDataDelete:   "数据删除",
	LogOpTypeDeviceReport: "设备上报",
	LogOpTypeIPRefresh:    "IP刷新",
}

// GetLogOpTypes 返回所有操作类型
var GetLogOpTypes = []string{
	LogOpTypeDataCreate,
	LogOpTypeDataUpdate,
	LogOpTypeDataDelete,
	LogOpTypeDeviceReport,
	LogOpTypeIPRefresh,
}

// --------

const (
	LogObjTypePlatformConfig       string = "platform_config"        // 平台配置
	LogObjTypeCdnRecord            string = "cdn_record"             // CDN记录
	LogObjTypeDnsRecord            string = "dns_record"             // DNS记录
	LogObjTypeDevice               string = "device"                 // 设备
	LogObjTypeDeviceIPRecord       string = "device_ip_report"       // 设备IP记录
	LogObjTypeIPGroup              string = "ip_group"               // IP组
	LogObjTypeIPGroupDevice        string = "ip_group_device"        // IP组设备
	LogObjTypeWebhook              string = "webhook"                // 通知
	LogObjTypeWebhookSubscription  string = "webhook_subscription"   // 通知订阅
	LogObjTypeDevicePortForward    string = "device_port_forward"    // 端口转发
	LogObjTypeDomainMonitor        string = "domain_monitor"         // 域名监控
	LogObjTypeDomainMonitorHistory string = "domain_monitor_history" // 域名监控历史记录
)

// GetLogObjTypeDescriptions 对象类型说明映射
var GetLogObjTypeDescriptions = map[string]string{
	LogObjTypePlatformConfig:       "平台配置",
	LogObjTypeCdnRecord:            "CDN记录",
	LogObjTypeDnsRecord:            "DNS记录",
	LogObjTypeDevice:               "设备",
	LogObjTypeDeviceIPRecord:       "设备IP记录",
	LogObjTypeIPGroup:              "IP组",
	LogObjTypeIPGroupDevice:        "IP组设备",
	LogObjTypeWebhook:              "通知",
	LogObjTypeWebhookSubscription:  "通知订阅",
	LogObjTypeDevicePortForward:    "端口转发",
	LogObjTypeDomainMonitor:        "域名监控",
	LogObjTypeDomainMonitorHistory: "域名监控历史记录",
}

// GetLogObjTypes 返回所有对象类型
var GetLogObjTypes = []string{
	LogObjTypePlatformConfig,
	LogObjTypeCdnRecord,
	LogObjTypeDnsRecord,
	LogObjTypeDevice,
	LogObjTypeDeviceIPRecord,
	LogObjTypeIPGroup,
	LogObjTypeIPGroupDevice,
	LogObjTypeWebhook,
	LogObjTypeWebhookSubscription,
	LogObjTypeDevicePortForward,
	LogObjTypeDomainMonitor,
	LogObjTypeDomainMonitorHistory,
}

// --------

const (
	LogStatusSuccess string = "success" // 成功
	LogStatusFailed  string = "failed"  // 失败
)

// GetLogStatusDescriptions 状态说明映射
var GetLogStatusDescriptions = map[string]string{
	LogStatusSuccess: "成功",
	LogStatusFailed:  "失败",
}

// GetLogStatuss 返回所有状态
var GetLogStatuss = []string{
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
