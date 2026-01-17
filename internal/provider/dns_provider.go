package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"
	webhookService "vibe_ip/internal/backend/webhook"
	"vibe_ip/internal/ent"
	"vibe_ip/internal/ent/webhooksubscription"
	"vibe_ip/internal/utils"
)

// ---------- DNS 模块 ----------

// DNSProvider 定义了 DNS 操作的接口
type DNSProvider interface {
	// Name 返回 DNS 提供程序的名称
	Name() (name string)
	// Check 检查 DNS 记录是否存在
	Check(ctx context.Context, domain string, subdomain string, originValue string, recordType string) (recordID string, isContinue bool, err error)
	// Update 更新 DNS 记录
	Update(ctx context.Context, recordID string, domain string, subdomain string, originValue string, recordType string) (err error)
	// ParseCustomConfig 解析 DNS 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// DNSTask 定义了 DNS 更新任务
type DNSTask struct {
	logger   *slog.Logger // 日志记录器
	provider DNSProvider  // DNS 提供程序

	config         *DNSConfigParams // 配置
	isNotification bool             // 是否通知

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置
}

// NewDNSTask 创建 DNS 更新任务
func NewDNSTask(logger *slog.Logger, provider DNSProvider, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *DNSTask, err error) {
	t = &DNSTask{
		logger:               logger,
		provider:             provider,
		config:               &DNSConfigParams{},
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}
	if len(t.platform_config_json) > 0 {
		var platformConfig PlatformConfigJson
		if err := json.Unmarshal(t.platform_config_json, &platformConfig); err != nil {
			t.logger.Error("解析平台配置失败",
				slog.String("提供商", t.provider.Name()),
				slog.Any("err", err),
			)
			return nil, err
		}
		t.config.domain = platformConfig.DomainName
	}
	return t, nil
}

type DNSConfigParams struct {
	domain    string // 域名名称
	subdomain string // 子域名

	recordID       string // 域名ID
	recordType     string // 记录类型
	oldOriginValue string // 旧记录值
	newOriginValue string // 新记录值
}

// Config 配置 DNS 参数
func (t *DNSTask) Config(subdomain string, recordType string, newOriginValue string) *DNSTask {
	t.config.subdomain = subdomain
	t.config.recordType = recordType
	t.config.newOriginValue = newOriginValue
	return t
}

// Validate 验证 DNS 更新任务配置
func (t *DNSTask) Validate() error {
	if t.config.domain == "" {
		return fmt.Errorf("domain 不能为空")
	}
	if t.config.subdomain == "" {
		return fmt.Errorf("subdomain 不能为空")
	}
	if t.config.recordType == "" {
		return fmt.Errorf("recordType 不能为空")
	}
	if t.config.newOriginValue == "" {
		return fmt.Errorf("newOriginValue 不能为空")
	}
	return nil
}

// Execute 执行 DNS 更新
func (t *DNSTask) Execute(ctx context.Context) (err error) {
	if t.provider == nil {
		return nil
	}
	t.logger.InfoContext(ctx, "开始执行",
		slog.String("提供商", t.provider.Name()),
		slog.String("域名名称", t.config.domain),
		slog.String("子域名", t.config.subdomain),
		slog.String("记录类型", t.config.recordType),
		slog.String("新记录值", t.config.newOriginValue),
		slog.String("自定义配置", string(t.record_config_json)),
	)
	defer t.logger.InfoContext(ctx, "执行完成")

	// 验证配置
	t.logger.DebugContext(ctx, "执行验证配置")
	if err := t.Validate(); err != nil {
		t.logger.ErrorContext(ctx, "验证配置失败",
			slog.String("提供商", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	if len(t.record_config_json) > 0 {
		t.logger.DebugContext(ctx, "执行解析自定义配置")
		if err := t.provider.ParseCustomConfig(t.record_config_json); err != nil {
			t.logger.ErrorContext(ctx, "解析自定义配置失败",
				slog.String("提供商", t.provider.Name()),
				slog.Any("err", err),
			)
			return err
		}
	}

	t.logger.DebugContext(ctx, "执行检查")
	var isContinue = true
	t.config.recordID, isContinue, err = t.provider.Check(
		ctx,
		t.config.domain,
		t.config.subdomain,
		t.config.newOriginValue,
		t.config.recordType,
	)
	if err != nil {
		t.logger.ErrorContext(ctx, "检查失败",
			slog.String("提供商", t.provider.Name()),
			slog.String("域名ID", t.config.recordID),
			slog.Any("err", err),
		)
		return err
	}

	if !isContinue {
		t.logger.WarnContext(ctx, "没有变化无需更新")
		return nil
	}

	t.logger.DebugContext(ctx, "执行更新")
	if err := t.provider.Update(
		ctx,
		t.config.recordID,
		t.config.domain,
		t.config.subdomain,
		t.config.newOriginValue,
		t.config.recordType,
	); err != nil {
		t.logger.ErrorContext(ctx, "更新失败",
			slog.String("提供商", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	// 可以下发通知
	t.isNotification = true

	return nil
}

// SaveRefresh 保存 DNS 记录刷新
func (t *DNSTask) SaveRefresh(ctx context.Context, db *ent.Client, id int64) (err error) {
	if db == nil || id == 0 {
		return nil
	}

	err = db.DNSRecord.UpdateOneID(id).
		SetRecordID(t.config.recordID).
		SetLastRefreshAt(time.Now()).
		Exec(ctx)
	if err != nil {
		t.logger.ErrorContext(ctx, "保存刷新失败",
			slog.String("提供商", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}
	return nil
}

// SaveResponse 保存 DNS 记录响应
func (t *DNSTask) SaveResponse(ctx context.Context, db *ent.Client, id int64, last_api_response string) (err error) {
	if db == nil || id == 0 {
		return nil
	}

	err = db.DNSRecord.UpdateOneID(id).
		SetLastAPIResponse(last_api_response).
		SetLastAPIResponseAt(time.Now()).
		Exec(ctx)
	if err != nil {
		t.logger.ErrorContext(ctx, "保存响应失败",
			slog.String("提供商", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}
	return nil
}

// Notify 执行 DNS 通知
func (t *DNSTask) Notify(ctx context.Context, webhookService *webhookService.Send, id int64) (err error) {
	if webhookService == nil || id == 0 || !t.isNotification {
		return nil
	}

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventTitle:     "DNS记录更新",
		EventType:      "DNS_UPDATE",
		EventTime:      time.Now().Format("2006-01-02 15:04:05"),
		EventBody:      "DNS记录更新事件",
		RecordDomain:   fmt.Sprintf("%s.%s", t.config.subdomain, t.config.domain),
		RecordType:     t.config.recordType,
		RecordOldValue: t.config.oldOriginValue,
		RecordNewValue: t.config.newOriginValue,
	}

	// 发送通知
	if err := webhookService.Send(
		ctx,
		webhooksubscription.ObjTypeDNSRecord,
		id,
		notifyData,
	); err != nil {
		t.logger.ErrorContext(ctx, "通知失败",
			slog.String("提供商", t.provider.Name()),
			slog.Any("err", err),
		)
	}

	return nil
}
