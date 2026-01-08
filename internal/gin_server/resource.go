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

	"github.com/denisbrodbeck/machineid"
	"github.com/dromara/carbon/v2"
	"github.com/gin-contrib/multitemplate"
)

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
			return m.config.Debug
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
		// pagination 组件
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"div":   func(a, b int) int { return a / b },
		"toInt": func(i int64) int { return int(i) },
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
		"fieldEnumMap": func(fieldType string) (enumItems []EnumItem) {
			switch fieldType {
			case "cdn_type":
				// CDN记录 记录类型
				for _, value := range schematype.GetCDNRecordTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetCDNRecordTypeDescriptions[value],
						Value: value,
					})
				}
			case "cdn_origin_protocol":
				// CDN记录 回源协议
				for _, value := range schematype.GetOriginProtocols {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetOriginProtocolDescriptions[value],
						Value: value,
					})
				}
			case "dns_type":
				// DNS记录 记录类型
				for _, value := range schematype.GetDNSRecordTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDNSRecordTypeDescriptions[value],
						Value: value,
					})
				}
			case "domain_monitor_check_protocol":
				// 域名监控 检查类型
				for _, value := range schematype.GetDomainMonitorCheckProtocols {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDomainMonitorCheckProtocolDescriptions[value],
						Value: value,
					})
				}
			case "domain_monitor_current_status":
				// 域名监控 当前运行状态
				for _, value := range schematype.GetDomainMonitorStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetDomainMonitorStatusDescriptions[value],
						Value: value,
					})
				}
			case "platform_config_capability":
				// 平台配置 平台支持的服务能力
				for _, value := range schematype.GetConfigCapabilitys {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetConfigCapabilityDescriptions[value],
						Value: value,
					})
				}
			case "platform_config_platform":
				// 平台配置 平台类型
				for _, value := range schematype.GetConfigPlatforms {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetConfigPlatformDescriptions[value],
						Value: value,
					})
				}
			case "ssl_record_status":
				// SSL记录 证书状态
				for _, value := range schematype.GetSSLRecordStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordStatusDescriptions[value],
						Value: value,
					})
				}
			case "ssl_record_provider":
				// SSL记录 CA提供商
				for _, value := range schematype.GetSSLRecordProviders {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordProviderDescriptions[value],
						Value: value,
					})
				}
			case "webhook_platform":
				// 通知 通知平台类型
				for _, value := range schematype.GetWebhookPlatforms {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetWebhookPlatformDescriptions[value],
						Value: value,
					})
				}
			case "webhook_subscription_obj_type":
				// 通知订阅 对象类型
				for _, value := range schematype.GetWebhookSubscriptionObjTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetWebhookSubscriptionObjTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_source":
				// 操作日志 操作来源
				for _, value := range schematype.GetLogSources {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogSourceDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_op_type":
				// 操作日志 操作类型
				for _, value := range schematype.GetLogOpTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogOpTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_obj_type":
				// 操作日志 对象类型
				for _, value := range schematype.GetLogObjTypes {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogObjTypeDescriptions[value],
						Value: value,
					})
				}
			case "operation_log_status":
				// 操作日志 状态
				for _, value := range schematype.GetLogStatuss {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetLogStatusDescriptions[value],
						Value: value,
					})
				}
			case "ssl_certificate_history_provider":
				// SSL证书申请历史记录 CA提供商
				for _, value := range schematype.GetSSLRecordProviders {
					enumItems = append(enumItems, EnumItem{
						Label: schematype.GetSSLRecordProviderDescriptions[value],
						Value: value,
					})
				}
			}
			return enumItems
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
		// 将任意类型转换为 JSON 字符串
		"marshalJSON": func(v any) string {
			if v == nil || v == "" {
				return ""
			}
			b, _ := json.Marshal(v)
			return string(b)
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
