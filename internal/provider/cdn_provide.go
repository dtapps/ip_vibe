package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// ---------- CDN 模块 ----------

// CDNProviderInternalData 定义了 CDN 提供程序的内部数据
type CDNProviderInternalData struct {
	RecordID   string // 域名ID
	CdnDomain  string // CDN域名
	IsContinue bool   // 是否继续
	Tips       string // 提示信息
}

// CDNProvider 定义 CDN 提供程序接口
type CDNProvider interface {
	// Name 返回 CDN 提供程序的名称
	Name() string
	// Check 检查 CDN 记录是否存在
	Check(ctx context.Context, domain string, subdomain string, originValues []string, recordType string) (data *CDNProviderInternalData, err error)
	// Update 更新 CDN 记录
	Update(ctx context.Context, domain string, subdomain string, originValues []string, recordType string) (err error)
	// ParseCustomConfig 解析 CDN 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// CDNTask 定义了 CDN 更新任务
type CDNTask struct {
	logger   *slog.Logger // 日志记录器
	provider CDNProvider  // CDN 提供程序

	config         *CDNConfigParams // 配置
	isNotification bool             // 是否通知

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置
}

func NewCDNTask(logger *slog.Logger, provider CDNProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *CDNTask, err error) {
	t = &CDNTask{
		logger:               logger,
		provider:             provider,
		config:               &CDNConfigParams{},
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}

	// 优化日志
	t.logger = t.logger.With(
		slog.String("模块", "CDN"),
		slog.String("提供商", t.provider.Name()),
	)

	if len(t.platform_config_json) > 0 {
		var platformConfig PlatformConfigJson
		if err := json.Unmarshal(t.platform_config_json, &platformConfig); err != nil {
			t.logger.Error("解析平台配置失败",
				slog.Any("err", err),
			)
			return nil, err
		}
		t.config.domain = platformConfig.DomainName
	}

	return t, nil
}

type CDNConfigParams struct {
	domain    string // 域名名称
	subdomain string // 子域名

	recordID        string   // 域名ID
	recordType      string   // 记录类型
	cdnDomain       string   // CDN域名
	oldOriginValues []string // 旧记录值列表
	newOriginValues []string // 新记录值列表
}

// Config 配置 DNS 参数
func (t *CDNTask) Config(subdomain string, recordType string, newOriginValues []string) *CDNTask {
	t.config.subdomain = subdomain
	t.config.recordType = recordType
	t.config.newOriginValues = newOriginValues
	return t
}

// Validate 验证 DNS 更新任务配置
func (t *CDNTask) Validate() error {
	if t.config.domain == "" {
		return fmt.Errorf("domain 不能为空")
	}
	if t.config.subdomain == "" {
		return fmt.Errorf("subdomain 不能为空")
	}
	if t.config.recordType == "" {
		return fmt.Errorf("recordType 不能为空")
	}
	if len(t.config.newOriginValues) == 0 {
		return fmt.Errorf("newOriginValues 不能为空")
	}
	return nil
}

// Execute 执行 CDN 更新
func (t *CDNTask) Execute(ctx context.Context) (err error) {
	if t.provider == nil {
		return nil
	}
	t.logger.InfoContext(ctx, "开始执行",
		slog.String("域名名称", t.config.domain),
		slog.String("子域名", t.config.subdomain),
		slog.String("记录类型", t.config.recordType),
		slog.String("新记录值列表", strings.Join(t.config.newOriginValues, ",")),
		slog.String("自定义配置", string(t.record_config_json)),
	)
	defer t.logger.InfoContext(ctx, "执行完成")

	// 验证配置
	t.logger.DebugContext(ctx, "执行验证配置")
	if err := t.Validate(); err != nil {
		t.logger.ErrorContext(ctx, "验证配置失败",
			slog.Any("err", err),
		)
		return err
	}

	if len(t.record_config_json) > 0 {
		t.logger.DebugContext(ctx, "执行解析自定义配置")
		if err := t.provider.ParseCustomConfig(t.record_config_json); err != nil {
			t.logger.ErrorContext(ctx, "解析自定义配置失败",
				slog.Any("err", err),
			)
			return err
		}
	}

	t.logger.DebugContext(ctx, "执行检查")
	checkData, err := t.provider.Check(
		ctx,
		t.config.domain,
		t.config.subdomain,
		t.config.newOriginValues,
		t.config.recordType,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "检查失败",
			slog.Any("err", err),
		)
		return err
	}
	t.config.recordID = checkData.RecordID
	t.config.cdnDomain = checkData.CdnDomain

	if !checkData.IsContinue {
		t.logger.WarnContext(ctx, "没有变化无需更新")
		return nil
	}

	t.logger.DebugContext(ctx, "执行更新",
		slog.String("域名ID", t.config.recordID),
		slog.String("CDN域名", t.config.cdnDomain),
		slog.String("提示", checkData.Tips),
	)
	if err := t.provider.Update(
		ctx,
		t.config.domain,
		t.config.subdomain,
		t.config.newOriginValues,
		t.config.recordType,
	); err != nil {
		t.logger.ErrorContext(ctx, "更新失败",
			slog.Any("err", err),
		)
		return err
	}

	// 可以下发通知
	t.isNotification = true

	return nil
}
