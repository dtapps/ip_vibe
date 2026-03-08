package gin_server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"vibe_ip/internal/ent/schema/schematype"
	"vibe_ip/internal/global"
	"vibe_ip/internal/helper/db_cache"
	"vibe_ip/internal/helper/sys_env"

	"github.com/denisbrodbeck/machineid"
	"github.com/dromara/carbon/v2"
	"github.com/dustin/go-humanize"
	"github.com/gin-contrib/multitemplate"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

type EnumDefinition struct {
	Values       []string
	Descriptions map[string]string // Value -> Label
}

var EnumRegistry = map[string]EnumDefinition{
	// CDN记录 记录类型
	"cdn_type": {
		Values:       schematype.GetCDNRecordTypes,
		Descriptions: schematype.GetCDNRecordTypeDescriptions,
	},
	// DNS记录 记录类型
	"dns_type": {
		Values:       schematype.GetDNSRecordTypes,
		Descriptions: schematype.GetDNSRecordTypeDescriptions,
	},
	// 域名监控 检查类型
	"domain_monitor_check_protocol": {
		Values:       schematype.GetDomainMonitorCheckProtocols,
		Descriptions: schematype.GetDomainMonitorCheckProtocolDescriptions,
	},
	// 域名监控 当前运行状态
	"domain_monitor_current_status": {
		Values:       schematype.GetDomainMonitorStatuss,
		Descriptions: schematype.GetDomainMonitorStatusDescriptions,
	},
	// 平台配置 平台支持的服务能力
	"platform_config_capability": {
		Values:       schematype.GetConfigCapabilitys,
		Descriptions: schematype.GetConfigCapabilityDescriptions,
	},
	// 平台配置 平台类型
	"platform_config_platform": {
		Values:       schematype.GetConfigPlatforms,
		Descriptions: schematype.GetConfigPlatformDescriptions,
	},
	// SSL记录 证书状态
	"ssl_record_status": {
		Values:       schematype.GetSSLRecordStatuss,
		Descriptions: schematype.GetSSLRecordStatusDescriptions,
	},
	// SSL记录 CA提供商
	"ssl_record_provider": {
		Values:       schematype.GetSSLRecordProviders,
		Descriptions: schematype.GetSSLRecordProviderDescriptions,
	},
	// 通知 通知平台类型
	"webhook_platform": {
		Values:       schematype.GetWebhookPlatforms,
		Descriptions: schematype.GetWebhookPlatformDescriptions,
	},
	// 通知订阅 对象类型
	"webhook_subscription_obj_type": {
		Values:       schematype.GetWebhookSubscriptionObjTypes,
		Descriptions: schematype.GetWebhookSubscriptionObjTypeDescriptions,
	},
	// 操作日志 操作来源
	"operation_log_source": {
		Values:       schematype.GetLogSources,
		Descriptions: schematype.GetLogSourceDescriptions,
	},
	// 操作日志 操作类型
	"operation_log_op_type": {
		Values:       schematype.GetLogOpTypes,
		Descriptions: schematype.GetLogOpTypeDescriptions,
	},
	// 操作日志 对象类型
	"operation_log_obj_type": {
		Values:       schematype.GetLogObjTypes,
		Descriptions: schematype.GetLogObjTypeDescriptions,
	},
	// 操作日志 状态
	"operation_log_status": {
		Values:       schematype.GetLogStatuss,
		Descriptions: schematype.GetLogStatusDescriptions,
	},
	// SSL证书申请历史记录 CA提供商
	"ssl_certificate_history_provider": {
		Values:       schematype.GetSSLRecordProviders,
		Descriptions: schematype.GetSSLRecordProviderDescriptions,
	},
	// SSL导入 来源
	"ssl_import_source": {
		Values:       schematype.GetSSLImportSources,
		Descriptions: schematype.GetSSLImportSourceDescriptions,
	},
	// SSL导入 状态
	"ssl_import_status": {
		Values:       schematype.GetSSLImportStatuss,
		Descriptions: schematype.GetSSLImportStatusDescriptions,
	},
	// TUN记录 服务类型
	"tun_serve_type": {
		Values:       schematype.GetTUNRecordServeTypes,
		Descriptions: schematype.GetTUNRecordServeTypeDescriptions,
	},
	// Cloudflare 隧道 隧道状态
	"cloudflare_tunnel_status": {
		Values:       schematype.GetCloudflareTunnelStatuss,
		Descriptions: schematype.GetCloudflareTunnelStatusDescriptions,
	},
	// Cloudflare 隧道 隧道类型
	"cloudflare_tunnel_tun_type": {
		Values:       schematype.GetCloudflareTunnelTunTypes,
		Descriptions: schematype.GetCloudflareTunnelTunTypeDescriptions,
	},
	// Cloudflare 站点 Cloudflare上的区域状态
	"cloudflare_zone_status": {
		Values:       schematype.GetCloudflareZoneStatuss,
		Descriptions: schematype.GetCloudflareZoneStatusDescriptions,
	},
	// Cloudflare 站点 完整区域意味着DNS托管在Cloudflare上。部分区域通常是合作伙伴托管的区域或CNAME设置
	"cloudflare_zone_type": {
		Values:       schematype.GetCloudflareZoneTypes,
		Descriptions: schematype.GetCloudflareZoneTypeDescriptions,
	},
	// Cloudflare 站点记录 数据录类型
	"cloudflare_zone_record_type": {
		Values:       schematype.GetCloudflareZoneRecordTypes,
		Descriptions: schematype.GetCloudflareZoneRecordTypeDescriptions,
	},
	// EdgeOne 站点 站点加速区域
	"edgeone_zone_area": {
		Values:       schematype.GetEdgeOneZoneAreas,
		Descriptions: schematype.GetEdgeOneZoneAreaDescriptions,
	},
	// EdgeOne 站点 站点接入类型
	"edgeone_zone_type": {
		Values:       schematype.GetEdgeOneZoneTypes,
		Descriptions: schematype.GetEdgeOneZoneTypeDescriptions,
	},
	// EdgeOne 站点 状态
	"edgeone_zone_status": {
		Values:       schematype.GetEdgeOneZoneStatuss,
		Descriptions: schematype.GetEdgeOneZoneStatusDescriptions,
	},
	// EdgeOne 站点 CNAME 接入状态
	"edgeone_zone_cname_status": {
		Values:       schematype.GetEdgeOneZoneCnameStatuss,
		Descriptions: schematype.GetEdgeOneZoneCnameStatusDescriptions,
	},
	// EdgeOne 站点 展示状态
	"edgeone_zone_active_status": {
		Values:       schematype.GetEdgeOneZoneActiveStatuss,
		Descriptions: schematype.GetEdgeOneZoneActiveStatusDescriptions,
	},
	// EdgeOne 站点 锁定状态
	"edgeone_zone_lock_status": {
		Values:       schematype.GetEdgeOneZoneLockStatuss,
		Descriptions: schematype.GetEdgeOneZoneLockStatusDescriptions,
	},
	// EdgeOne 站点源站组 源站组类型
	"edgeone_zone_origin_group_type": {
		Values:       schematype.GetEdgeOneZoneOriginGroupTypes,
		Descriptions: schematype.GetEdgeOneZoneOriginGroupTypeDescriptions,
	},
	// EdgeOne 站点DNS记录 DNS 记录类型
	"edgeone_zone_dns_record_type": {
		Values:       schematype.GetEdgeOneZoneDNSRecordTypes,
		Descriptions: schematype.GetEdgeOneZoneDNSRecordTypeDescriptions,
	},
	// EdgeOne 站点DNS记录 DNS 记录解析状态
	"edgeone_zone_dns_record_status": {
		Values:       schematype.GetEdgeOneZoneDNSRecordStatuss,
		Descriptions: schematype.GetEdgeOneZoneDNSRecordStatusDescriptions,
	},
	// EdgeOne 站点加速域名 加速域名状态
	"edgeone_zone_acceleration_domain_domain_status": {
		Values:       schematype.GetEdgeOneZoneAccelerationDomainDomainStatuss,
		Descriptions: schematype.GetEdgeOneZoneAccelerationDomainDomainStatusDescriptions,
	},
	// EdgeOne 站点加速域名 IPv6 状态
	"edgeone_zone_acceleration_domain_ipv6_status": {
		Values:       schematype.GetEdgeOneZoneAccelerationDomainIPv6Statuss,
		Descriptions: schematype.GetEdgeOneZoneAccelerationDomainIPv6StatusDescriptions,
	},
	// EdgeOne 站点加速域名 加速域名归属权验证状态
	"edgeone_zone_acceleration_domain_identification_status": {
		Values:       schematype.GetEdgeOneZoneAccelerationDomainIdentificationStatuss,
		Descriptions: schematype.GetEdgeOneZoneAccelerationDomainIdentificationStatusDescriptions,
	},
	// EdgeOne 站点加速域名 源站信息 源站类型
	"edgeone_zone_acceleration_domain_origin_detail_origin_type": {
		Values:       schematype.GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypes,
		Descriptions: schematype.GetEdgeOneZoneAccelerationDomainOriginDetailOriginTypeDescriptions,
	},
	// EdgeOne 站点加速域名 回源协议
	"edgeone_zone_acceleration_domain_origin_protocol": {
		Values:       schematype.GetEdgeOneZoneAccelerationDomainOriginProtocols,
		Descriptions: schematype.GetEdgeOneZoneAccelerationDomainOriginProtocolDescriptions,
	},
	// ESA 站点 站点接入类型
	"esa_site_access_type": {
		Values:       schematype.GetEsaSiteAccessTypes,
		Descriptions: schematype.GetEsaSiteAccessTypeDescriptions,
	},
	// ESA 站点 站点加速区域
	"esa_site_coverage": {
		Values:       schematype.GetEsaSiteCoverages,
		Descriptions: schematype.GetEsaSiteCoverageDescriptions,
	},
	// ESA 站点 站点状态
	"esa_site_status": {
		Values:       schematype.GetEsaSiteStatuss,
		Descriptions: schematype.GetEsaSiteStatusDescriptions,
	},
	// ESA 站点记录 记录加速时的业务场景
	"esa_site_record_biz_name": {
		Values:       schematype.GetEsaSiteRecordBizNames,
		Descriptions: schematype.GetEsaSiteRecordBizNameDescriptions,
	},
	// ESA 站点记录 CNAME 记录的源站类型
	"esa_site_record_record_source_type": {
		Values:       schematype.GetEsaSiteRecordRecordSourceTypes,
		Descriptions: schematype.GetEsaSiteRecordRecordSourceTypeDescriptions,
	},
	// ESA 站点记录 记录的 DNS 类型
	"esa_site_record_record_type": {
		Values:       schematype.GetEsaSiteRecordRecordTypes,
		Descriptions: schematype.GetEsaSiteRecordRecordTypeDescriptions,
	},
	// ESA 站点记录 回源 HOST 策略
	"esa_site_record_host_policy": {
		Values:       schematype.GetEsaSiteRecordHostPolicys,
		Descriptions: schematype.GetEsaSiteRecordHostPolicyDescriptions,
	},
	// 定时任务 任务类型
	"cron_job_task_type": {
		Values:       schematype.GetCronJobTaskTypes,
		Descriptions: schematype.GetCronJobTaskTypeDescriptions,
	},
	// 定时任务 运行类型
	"cron_job_job_type": {
		Values:       schematype.GetCronJobJobTypes,
		Descriptions: schematype.GetCronJobJobTypeDescriptions,
	},
	// 定时任务 任务分类
	"cron_job_job_category": {
		Values:       schematype.GetCronJobJobCategorys,
		Descriptions: schematype.GetCronJobJobCategoryDescriptions,
	},
	// 定时任务 业务类型
	"cron_job_business_type": {
		Values:       schematype.GetCronJobBusinessTypes,
		Descriptions: schematype.GetCronJobBusinessTypeDescriptions,
	},
	// 定时任务 间隔时间单位
	"cron_job_interval_unit": {
		Values:       schematype.GetCronJobIntervalUnits,
		Descriptions: schematype.GetCronJobIntervalUnitDescriptions,
	},
	// 定时任务 状态
	"cron_job_status": {
		Values:       schematype.GetCronJobStatuss,
		Descriptions: schematype.GetCronJobStatusDescriptions,
	},
	// 定时任务日志 状态
	"cron_job_log_status": {
		Values:       schematype.GetCronJobLogStatuss,
		Descriptions: schematype.GetCronJobLogStatusDescriptions,
	},
}

// 将原 main.go 的 funcMap 逻辑移入此处
func (m *GinServerManager) getTemplateFuncMap() template.FuncMap {

	// EnumItem 表示单个枚举项
	type EnumItem struct {
		Label string
		Value string
	}

	return template.FuncMap{
		// 判断是否为调试模式
		"isDebug": func() bool {
			return sys_env.GvaEnv.DEBUG
		},
		// 设备离线开启状态
		"deviceOfflineEnabled": func() bool {
			if db_cache.GvaSetting.DeviceOfflineInvalidTime <= 0 {
				return false
			}
			return true
		},
		// 格式化时间为字符串，若为空则返回 "—"
		"formatTimeOrNil": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return t.Format(time.DateTime)
		},
		// 格式化时间为 "X 前" 格式，如 "2 小时前"
		"diffForHumans": func(t time.Time) string {
			if t.IsZero() {
				return "—"
			}
			return carbon.NewCarbon(t).DiffForHumans()
		},
		// 格式化时间为 "HH:mm:ss" 格式，如 "14:30:00"
		// "formatTime": func(t time.Time) string {
		// 	return t.Format("15:04:05")
		// },
		// 格式化字节为人类可读格式，如 "1.23 MB"
		"formatBytes": func(b int64) string {
			return humanize.Bytes(uint64(b))
		},
		// 格式化数字为人类可读格式，如 "1,234"
		"formatNumber": func(b int64) string {
			return humanize.Comma(b)
		},
		// 格式化字节率为人类可读格式，如 "1.23 KB/s"
		"formatBps": func(b int64) string {
			// 使用 SI 标准 (1000进制)
			const unit = 1000
			if b < unit {
				return fmt.Sprintf("%d bps", b)
			}
			div, exp := int64(unit), 0
			for n := b / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			// "kMGTPE"[exp] 对应 K, M, G, T...
			return fmt.Sprintf("%.2f %cbps", float64(b)/float64(div), "kMGTPE"[exp])
		},
		// 格式化毫秒为人类可读格式，如 "1.23 s" 或 "123.45 ms"
		"formatMs": func(ms float64) string {
			if ms >= 1000 {
				return fmt.Sprintf("%.2f s", ms/1000)
			}
			return fmt.Sprintf("%.2f ms", ms)
		},
		// 字符串去空格
		"trim": strings.TrimSpace,
		// 字符串转换为大写
		"upper": strings.ToUpper,
		// 字符串替换
		"replace": strings.Replace,
		// 字符串替换所有
		"replaceAll": strings.ReplaceAll,
		// JSON 人性化
		"jsonify": func(v any) template.HTML {
			jsonStr, err := json.Marshal(v)
			if err != nil {
				m.logger.Error("[jsonify] 序列化 JSON 失败",
					slog.Any("err", err),
				)
				return template.HTML("—")
			}
			return template.HTML(strings.ReplaceAll(string(jsonStr), "\n", ""))
		},
		// 获取当前设备 Code
		"selfDeviceCode": func() string {
			code, err := machineid.ID()
			if err != nil {
				m.logger.Error("[selfDeviceCode] 获取当前设备 Code 失败",
					slog.Any("err", err),
				)
				return "—"
			}
			return code
		},
		// 分页组件
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"div": func(a, b int) int {
			return a / b
		},
		// 转换类型
		"toInt": func(i any) int {
			return cast.ToInt(i)
		},
		"toInt64": func(i any) int64 {
			return cast.ToInt64(i)
		},
		"toUint": func(i any) uint {
			return cast.ToUint(i)
		},
		"toUint64": func(i any) uint64 {
			return cast.ToUint64(i)
		},
		// switch / icons / button 组件
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				err := fmt.Errorf("期望有偶数个参数")
				m.logger.Error("[dict] 函数参数错误",
					slog.Any("err", err),
				)
				return nil, err
			}
			ma := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					err := fmt.Errorf("键必须是字符串")
					m.logger.Error("[dict] 函数参数错误",
						slog.Any("err", err),
					)
					return nil, err
				}
				ma[key] = values[i+1]
			}
			return ma, nil
		},
		// 获取字段的枚举项
		"fieldEnumMap": func(fieldType string) []EnumItem {
			def, exists := EnumRegistry[fieldType]
			if !exists {
				return nil // 或返回空 slice：return []EnumItem{}
			}
			enumItems := make([]EnumItem, 0, len(def.Values))
			for _, value := range def.Values {
				label := def.Descriptions[value]
				if label == "" {
					label = value // 容错：若无描述，用值本身当标签
				}
				enumItems = append(enumItems, EnumItem{
					Label: label,
					Value: value,
				})
			}
			return enumItems
		},
		// 获取字段的枚举项值
		"fieldEnumValues": func(fieldType string) []string {
			if def, ok := EnumRegistry[fieldType]; ok {
				// 返回副本
				values := make([]string, len(def.Values))
				copy(values, def.Values)
				return values
			}
			return []string{}
		},
		// 获取字段枚举值对应的中文标签
		"fieldEnumLabel": func(fieldType string, value any) string {
			valueStr := cast.ToString(value)
			if def, ok := EnumRegistry[fieldType]; ok {
				if label, exists := def.Descriptions[valueStr]; exists {
					return label
				}
			}
			// 如果没找到描述，返回原值
			return valueStr
		},
		// 将多个字符串参数拼接为一个字符串
		"join": func(s ...string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.Join(s, "")
		},
		// 将字符串按指定分隔符分割为字符串切片
		"slice": func(s ...string) []string {
			if len(s) == 0 {
				return []string{}
			}
			return s
		},
		// 将字符串按指定分隔符分割为字符串切片
		"split": func(s, sep string) []string {
			if s == "" || sep == "" {
				return []string{}
			}
			return strings.Split(s, sep)
		},
		// 判断字符串是否包含指定子字符串
		"contains": func(s, substr string) bool {
			if s == "" || substr == "" {
				return false
			}
			return strings.Contains(s, substr)
		},
		// 判断字符串是否相等（不区分大小写）
		"equalFold": func(s, t string) bool {
			return strings.EqualFold(s, t)
		},
		// 从 JSON 中提取字段
		"jsonGet": func(data json.RawMessage, path string) any {
			result := gjson.GetBytes(data, path)
			if !result.Exists() {
				m.logger.Debug("[jsonGet] 字段不存在",
					slog.String("path", path),
				)
				return nil
			}
			return result.Value()
		},
		// 将 JSON 转换为 map[string]any
		"jsonToMap": func(data json.RawMessage) map[string]any {
			var result map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				m.logger.Error("[jsonToMap] 反序列化 JSON 失败",
					slog.Any("err", err),
				)
				return nil
			}
			return result
		},
		// 将 JSON 转换为 []map[string]any
		"jsonToSlice": func(data json.RawMessage) []map[string]any {
			var result []map[string]any
			if err := json.Unmarshal(data, &result); err != nil {
				m.logger.Error("[jsonToSlice] 反序列化 JSON 失败",
					slog.Any("err", err),
				)
				return nil
			}
			return result
		},
		// 将任意类型转换为 JSON 字符串
		"marshalJSON": func(v any) string {
			if v == nil || v == "" {
				return ""
			}
			b, err := json.Marshal(v)
			if err != nil {
				m.logger.Error("[marshalJSON] 序列化 JSON 失败",
					slog.Any("err", err),
				)
				return ""
			}
			return string(b)
		},
		// 解析 解析记录的 ttl
		"parseRecordTTL": func(ttlType string, ttlValue any) string {
			// 尝试将 ttlValue 转换为整数
			ttl := cast.ToInt(ttlValue)
			// 根据 ttlType 进行解析
			if ttlType == "cloudflare-zone-record" || ttlType == "esa-site-record" {
				if ttl == 1 {
					return "自动"
				}
				switch {
				case ttl < 60:
					return fmt.Sprintf("%d 秒", ttl)
				case ttl < 3600:
					return fmt.Sprintf("%d 分钟", ttl/60)
				case ttl < 86400:
					return fmt.Sprintf("%d 小时", ttl/3600)
				default:
					return fmt.Sprintf("%d 天", ttl/86400)
				}
			}
			return fmt.Sprintf("%v", ttlValue)
		},
		// 获取缓存键值
		"getCacheKey": func(key string, defaultValue string) string {
			return db_cache.GetCacheKey(context.Background(), global.GvaEntDb, key, defaultValue)
		},
	}
}

// InitResource 整合了原 main.go 的 web_resource 逻辑
func (m *GinServerManager) InitResource(ctx context.Context, embeddedFiles embed.FS) {
	funcMap := m.getTemplateFuncMap()

	// 判断是否为开发模式
	devMode := m.config.Version == "dev"
	if devMode {
		// 开发环境：读取本地磁盘
		m.engine.HTMLRender = m.loadTemplates(m.logger, "cmd/backend/templates", funcMap)
		m.engine.StaticFS("/static", http.Dir("cmd/backend/static"))
	} else {
		// 生产环境：读取 embed.FS
		m.engine.HTMLRender = m.loadTemplatesFS(m.logger, embeddedFiles, funcMap)
		staticFS, _ := fs.Sub(embeddedFiles, "static")
		m.engine.StaticFS("/static", http.FS(staticFS))
	}
}

// 开发环境
func (m *GinServerManager) loadTemplates(logger *slog.Logger, templatePath string, funcMap template.FuncMap) multitemplate.Renderer {

	r := multitemplate.NewRenderer()

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	logger.Debug("开发环境",
		slog.String("layout", layout),
		slog.String("index", index),
	)

	// 所有业务页面
	pages, err := filepath.Glob(filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		logger.Error("开发环境",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := filepath.Glob(filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		logger.Error("开发环境",
			slog.Any("err", err),
		)
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		logger.Debug("开发环境",
			slog.String("components", name),
		)
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := os.Stat(path)
		// if err != nil {
		// 	logger.Error("开发环境",
		// 		slog.Any("err", err),
		// 	)
		// }
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		logger.Debug("开发环境",
			slog.String("pages", name),
		)

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html 和 _view.html 和 _filter.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_filter.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件
		files = append(files, p)             // 当前页面
		logger.Debug("开发环境",
			slog.String("page", strings.Join(files, " -> ")),
		)

		r.AddFromFilesFuncs(
			name,
			funcMap,
			files...,
		)
	}

	return r
}

// 生产环境
func (m *GinServerManager) loadTemplatesFS(logger *slog.Logger, efs embed.FS, funcMap template.FuncMap) multitemplate.Renderer {

	r := multitemplate.NewRenderer()

	templatePath := "templates"

	layout := filepath.Join(templatePath, "layout", "base.html")
	index := filepath.Join(templatePath, "pages", "index.html")
	logger.Debug("生产环境",
		slog.String("layout", layout),
		slog.String("index", index),
	)

	// 所有业务页面
	pages, err := fs.Glob(efs, filepath.Join(templatePath, "pages", "*.html"))
	if err != nil {
		logger.Error("生产环境",
			slog.Any("err", err),
		)
		panic(err)
	}

	// 所有组件（自动注入）
	components, err := fs.Glob(efs, filepath.Join(templatePath, "components", "*.html"))
	if err != nil {
		logger.Error("生产环境",
			slog.Any("err", err),
		)
		panic(err)
	}
	for _, c := range components {
		name := filepath.Join("components", filepath.Base(c))
		logger.Debug("生产环境",
			slog.String("components", name),
		)
	}

	// 检查文件是否存在
	var fileExists func(string) bool = func(path string) bool {
		_, err := efs.Open(path)
		// if err != nil {
		// 	logger.Error("生产环境",
		// 		slog.Any("err", err),
		// 	)
		// }
		return err == nil
	}

	for _, p := range pages {
		name := filepath.Join("pages", filepath.Base(p))
		logger.Debug("生产环境",
			slog.String("pages", name),
		)

		var extras []string

		var baseName string
		if strings.HasSuffix(p, "_detail.html") {
			// 特殊情况：如 p = "index_detail.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_detail.html")
		} else if strings.HasSuffix(p, "_history.html") {
			// 普通情况，如 p = "index_history.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, "_history.html")
		} else {
			// 普通情况，如 p = "index.html" → baseName = "index"
			baseName = strings.TrimSuffix(p, ".html")
		}

		// 检查对应的 _form.html 和 _popup.html 和 _view.html 和 _filter.html
		if formFile := baseName + "_form.html"; fileExists(formFile) {
			extras = append(extras, formFile)
		}
		if popupFile := baseName + "_popup.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_view.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}
		if popupFile := baseName + "_filter.html"; fileExists(popupFile) {
			extras = append(extras, popupFile)
		}

		// 页面模板 = base + index + components + extras + page
		files := []string{layout, index}
		files = append(files, components...) // 注入全部组件
		files = append(files, extras...)     // 注入额外文件

		files = append(files, p) // 当前页面
		logger.Debug("生产环境",
			slog.String("page", strings.Join(files, " -> ")),
		)

		r.AddFromFSFuncs(
			name,
			funcMap,
			efs,
			files...,
		)
	}

	return r
}
