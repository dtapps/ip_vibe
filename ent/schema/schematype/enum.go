package schematype

const (
	OriginProtocolHTTP   string = "http"   // HTTP协议
	OriginProtocolHTTPS  string = "https"  // HTTPS协议
	OriginProtocolFollow string = "follow" // 跟随协议
)

// GetOriginProtocol 返回回源协议
var GetOriginProtocol = func() []string {
	return []string{OriginProtocolHTTP, OriginProtocolHTTPS, OriginProtocolFollow}
}

const (
	RecordTypeA     string = "A"     // A记录
	RecordTypeAAAA  string = "AAAA"  // AAAA记录
	RecordTypeCNAME string = "CNAME" // CNAME记录
)

// GetRecordType 返回所有记录类型
var GetRecordType = func() []string {
	return []string{RecordTypeA, RecordTypeAAAA, RecordTypeCNAME}
}

// GetDNSRecordType 返回 DNS 记录类型
var GetDNSRecordType = func() []string {
	return []string{RecordTypeA, RecordTypeAAAA, RecordTypeCNAME}
}

// GetCDNRecordType 返回 CDN 记录类型
var GetCDNRecordType = func() []string {
	return []string{RecordTypeA, RecordTypeAAAA}
}

const (
	ConfigCapabilityDns string = "dns" // DNS服务能力
	ConfigCapabilityCdn string = "cdn" // CDN服务能力
	ConfigCapabilityAll string = "all" // 所有服务能力
)

// GetConfigCapability 返回所有服务能力
var GetConfigCapability = func() []string {
	return []string{ConfigCapabilityDns, ConfigCapabilityCdn, ConfigCapabilityAll}
}

const (
	ConfigPlatformCloudflare string = "cloudflare" // Cloudflare
	ConfigPlatformEdgeOne    string = "edgeone"    // EdgeOne
	ConfigPlatformDnspod     string = "dnspod"     // DNSPod
)

// GetConfigPlatform 返回所有平台类型
var GetConfigPlatform = func() []string {
	return []string{ConfigPlatformCloudflare, ConfigPlatformEdgeOne, ConfigPlatformDnspod}
}

const (
	WebhookPlatformWechatWork string = "wechat_work" // 企业微信
	WebhookPlatformDingtalk   string = "dingtalk"    // 钉钉
	WebhookPlatformFeishu     string = "feishu"      // 飞书
	WebhookPlatformBark       string = "bark"        // Bark
	WebhookPlatformServerChan string = "server_chan" // Server酱
	WebhookPlatformPushDeer   string = "push_deer"   // PushDeer
	WebhookPlatformCustomHTTP string = "custom_http" // 自定义 HTTP 平台
)

// GetWebhookPlatform 返回所有 Webhook 平台
var GetWebhookPlatform = func() []string {
	return []string{WebhookPlatformWechatWork, WebhookPlatformDingtalk, WebhookPlatformFeishu, WebhookPlatformBark, WebhookPlatformServerChan, WebhookPlatformPushDeer, WebhookPlatformCustomHTTP}
}

const (
	WebhookSubscriptionObjTypeCdnRecord string = "cdn_record" // CDN 记录
	WebhookSubscriptionObjTypeDnsRecord string = "dns_record" // DNS 记录
	WebhookSubscriptionObjTypeDevice    string = "device"     // 设备
)

// GetWebhookSubscriptionObjType 返回所有 Webhook 订阅对象类型
var GetWebhookSubscriptionObjType = func() []string {
	return []string{WebhookSubscriptionObjTypeCdnRecord, WebhookSubscriptionObjTypeDnsRecord, WebhookSubscriptionObjTypeDevice}
}

const (
	LogOpTypeDataCreate   string = "data_create"   // 数据创建
	LogOpTypeDataUpdate   string = "data_update"   // 数据更新
	LogOpTypeDataDelete   string = "data_delete"   // 数据删除
	LogOpTypeDeviceReport string = "device_report" // 设备上报
	LogOpTypeIPRefresh    string = "ip_refresh"    // IP 刷新
)

// GetLogOpType 返回所有操作类型
var GetLogOpType = func() []string {
	return []string{LogOpTypeDataCreate, LogOpTypeDataUpdate, LogOpTypeDataDelete, LogOpTypeDeviceReport, LogOpTypeIPRefresh}
}

const (
	LogObjTypePlatformConfig      string = "platform_config"      // 平台配置
	LogObjTypeCdnRecord           string = "cdn_record"           // CDN 记录
	LogObjTypeDnsRecord           string = "dns_record"           // DNS 记录
	LogObjTypeDevice              string = "device"               // 设备
	LogObjTypeDeviceIPRecord      string = "device_ip_report"     // 设备 IP 记录
	LogObjTypeIPGroup             string = "ip_group"             // IP 组
	LogObjTypeIPGroupDevice       string = "ip_group_device"      // IP 组设备
	LogObjTypeWebhook             string = "webhook"              // 通知
	LogObjTypeWebhookSubscription string = "webhook_subscription" // 通知订阅
)

// GetLogObjType 返回所有对象类型
var GetLogObjType = func() []string {
	return []string{LogObjTypePlatformConfig, LogObjTypeCdnRecord, LogObjTypeDnsRecord, LogObjTypeDevice, LogObjTypeDeviceIPRecord, LogObjTypeIPGroup, LogObjTypeIPGroupDevice, LogObjTypeWebhook, LogObjTypeWebhookSubscription}
}

const (
	LogStatusSuccess string = "success" // 成功
	LogStatusFailed  string = "failed"  // 失败
)

// GetLogStatus 返回所有状态
var GetLogStatus = func() []string {
	return []string{LogStatusSuccess, LogStatusFailed}
}
