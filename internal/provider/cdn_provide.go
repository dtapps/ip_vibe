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
	Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (data *CDNProviderInternalData, err error)
	// Update 更新 CDN 记录
	Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_values []string) (err error)
	// ParseCustomConfig 解析 CDN 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// CDNTask 定义了 CDN 更新任务
type CDNTask struct {
	logger   *slog.Logger // 日志记录器
	provider CDNProvider  // CDN 提供程序

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置

	isNotification bool // 是否通知
}

func NewCDNTask(logger *slog.Logger, provider CDNProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *CDNTask, err error) {
	t = &CDNTask{
		logger:               logger,
		provider:             provider,
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}

	// 优化日志
	t.logger = t.logger.With(
		slog.String("模块", "CDN"),
		slog.String("提供商", t.provider.Name()),
	)

	return t, nil
}

// CDNConfigParams 定义了 CDN 更新任务配置参数
type CDNConfigParams struct {
	DomainName   string   // 域名名称
	SubDomain    string   // 子域名
	FullDomain   string   // 完整域名
	RecordType   string   // 记录类型
	OriginValues []string // 源站列表
}

// Validate 验证 DNS 更新任务配置
func (t *CDNConfigParams) Validate() error {
	if t.DomainName == "" {
		return fmt.Errorf("DomainName 不能为空")
	}
	if t.SubDomain == "" {
		return fmt.Errorf("SubDomain 不能为空")
	}
	if t.FullDomain == "" {
		return fmt.Errorf("FullDomain 不能为空")
	}
	if t.RecordType == "" {
		return fmt.Errorf("RecordType 不能为空")
	}
	if len(t.OriginValues) == 0 {
		return fmt.Errorf("OriginValues 不能为空")
	}
	return nil
}

// Execute 执行 CDN 更新
func (t *CDNTask) Execute(ctx context.Context, params *CDNConfigParams) (err error) {
	if t.provider == nil {
		return nil
	}
	t.logger.InfoContext(ctx, "开始执行",
		slog.String("域名名称", params.DomainName),
		slog.String("子域名", params.SubDomain),
		slog.String("完整域名", params.FullDomain),
		slog.String("记录类型", params.RecordType),
		slog.String("源站列表", strings.Join(params.OriginValues, ",")),
		slog.String("自定义配置", string(t.record_config_json)),
	)
	defer t.logger.InfoContext(ctx, "执行完成")

	// 验证配置
	t.logger.DebugContext(ctx, "执行验证配置")
	err = params.Validate()
	if err != nil {
		t.logger.ErrorContext(ctx, "验证配置失败",
			slog.Any("err", err),
		)
		return err
	}

	if len(t.record_config_json) > 0 {
		t.logger.DebugContext(ctx, "执行解析自定义配置")
		err = t.provider.ParseCustomConfig(t.record_config_json)
		if err != nil {
			t.logger.ErrorContext(ctx, "解析自定义配置失败",
				slog.Any("err", err),
			)
			return err
		}
	}

	t.logger.DebugContext(ctx, "执行检查")
	checkData, err := t.provider.Check(
		ctx,
		params.DomainName,
		params.SubDomain,
		params.FullDomain,
		params.RecordType,
		params.OriginValues,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "检查失败",
			slog.Any("err", err),
		)
		return err
	}

	if !checkData.IsContinue {
		t.logger.WarnContext(ctx, "没有变化无需更新")
		return nil
	}

	t.logger.DebugContext(ctx, "执行更新",
		slog.String("域名ID", checkData.RecordID),
		slog.String("CDN域名", checkData.CdnDomain),
		slog.String("提示", checkData.Tips),
	)
	err = t.provider.Update(
		ctx,
		params.DomainName,
		params.SubDomain,
		params.FullDomain,
		params.RecordType,
		params.OriginValues,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "更新失败",
			slog.Any("err", err),
		)
		return err
	}

	// 可以下发通知
	t.isNotification = true

	return nil
}

// IsNotification 是否通知
func (t *CDNTask) IsNotification() bool {
	return t.isNotification
}
