package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

// ---------- DNS 模块 ----------

// DNSProviderInternalData 定义了 DNS 提供程序的内部数据
type DNSProviderInternalData struct {
	RecordID   string // 域名ID
	IsContinue bool   // 是否继续
	Tips       string // 提示信息
}

// DNSProvider 定义了 DNS 操作的接口
type DNSProvider interface {
	// Name 返回 DNS 提供程序的名称
	Name() (name string)
	// Check 检查 DNS 记录是否存在
	Check(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (data *DNSProviderInternalData, err error)
	// Update 更新 DNS 记录
	Update(ctx context.Context, domain_name string, sub_domain string, full_domain string, record_type string, origin_value string) (err error)
	// ParseCustomConfig 解析 DNS 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// DNSTask 定义了 DNS 更新任务
type DNSTask struct {
	logger   *slog.Logger // 日志记录器
	provider DNSProvider  // DNS 提供程序

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置

	isNotification bool   // 是否通知
	recordID       string // 域名ID
}

// NewDNSTask 创建 DNS 更新任务
func NewDNSTask(logger *slog.Logger, provider DNSProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *DNSTask, err error) {
	t = &DNSTask{
		logger:               logger,
		provider:             provider,
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}

	// 优化日志
	t.logger = t.logger.With(
		slog.String("模块", "DNS"),
		slog.String("提供商", t.provider.Name()),
	)

	return t, nil
}

// DNSConfigParams 定义了 DNS 更新任务的配置参数
type DNSConfigParams struct {
	DomainName  string // 域名名称
	SubDomain   string // 子域名
	FullDomain  string // 完整域名
	RecordType  string // 记录类型
	OriginValue string // 源站值
}

// Validate 验证 DNS 更新任务配置
func (t *DNSConfigParams) Validate() error {
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
	if t.OriginValue == "" {
		return fmt.Errorf("OriginValue 不能为空")
	}
	return nil
}

// Execute 执行 DNS 更新
func (t *DNSTask) Execute(ctx context.Context, config *DNSConfigParams) (err error) {
	if t.provider == nil {
		return nil
	}
	t.logger.InfoContext(ctx, "开始执行",
		slog.String("域名名称", config.DomainName),
		slog.String("子域名", config.SubDomain),
		slog.String("完整域名", config.FullDomain),
		slog.String("记录类型", config.RecordType),
		slog.String("记录值", config.OriginValue),
		slog.String("自定义配置", string(t.record_config_json)),
	)
	defer t.logger.InfoContext(ctx, "执行完成")

	// 验证配置
	err = config.Validate()
	if err != nil {
		t.logger.ErrorContext(ctx, "验证配置失败",
			slog.Any("err", err),
		)
		return err
	}

	// 解析 DNS 自定义配置
	if len(t.record_config_json) > 0 {
		err := t.provider.ParseCustomConfig(t.record_config_json)
		if err != nil {
			t.logger.ErrorContext(ctx, "解析自定义配置失败",
				slog.Any("err", err),
			)
			return err
		}
	}

	// 检查 DNS 记录是否存在
	checkData, err := t.provider.Check(
		ctx,
		config.DomainName,
		config.SubDomain,
		config.FullDomain,
		config.RecordType,
		config.OriginValue,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "检查失败",
			slog.Any("err", err),
		)
		return err
	}

	t.recordID = checkData.RecordID

	if !checkData.IsContinue {
		t.logger.WarnContext(ctx, "没有变化无需更新")
		return nil
	}

	t.logger.DebugContext(ctx, "执行更新",
		slog.String("域名ID", checkData.RecordID),
		slog.String("提示", checkData.Tips),
	)
	err = t.provider.Update(
		ctx,
		config.DomainName,
		config.SubDomain,
		config.FullDomain,
		config.RecordType,
		config.OriginValue,
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
func (t *DNSTask) IsNotification() bool {
	return t.isNotification
}

// GetRecordID 获取域名ID
func (t *DNSTask) GetRecordID() string {
	return t.recordID
}
