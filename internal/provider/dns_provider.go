package provider

import (
	"context"
	"encoding/json"
	"errors"
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
	// Check 检查 DNS 记录是否存在
	Check(ctx context.Context, domain string, subdomain string, recordType string) (recordID string, value string, err error)
	// Update 更新 DNS 记录
	Update(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error)
	// Save 保存 DNS 记录
	Save(ctx context.Context, recordID string, domain string, subdomain string, value string, recordType string) (err error)
	// Name 返回 DNS 提供程序的名称
	Name() (name string)
	// ParseCustomConfig 解析 DNS 自定义配置
	ParseCustomConfig(config_json json.RawMessage) (err error)
}

// DNSTask 定义了 DNS 更新任务
type DNSTask struct {
	logger    *slog.Logger // 日志记录器
	provider  DNSProvider  // DNS 提供程序
	domain    string       // 域名
	subdomain string       // 子域名

	platform_config_json json.RawMessage // 平台配置
	record_config_json   json.RawMessage // 记录配置
}

// NewDNSTask 创建 DNS 更新任务
func NewDNSTask(logger *slog.Logger, provider DNSProvider, subdomain string, platform_config_json json.RawMessage, record_config_json json.RawMessage) (t *DNSTask, err error) {
	t = &DNSTask{
		logger:               logger,
		provider:             provider,
		subdomain:            subdomain,
		platform_config_json: platform_config_json,
		record_config_json:   record_config_json,
	}
	if len(t.platform_config_json) > 0 {
		var platformConfig PlatformConfigJson
		if err := json.Unmarshal(t.platform_config_json, &platformConfig); err != nil {
			t.logger.Error("【DNS】解析平台配置失败",
				slog.String("name", t.provider.Name()),
				slog.Any("err", err),
			)
			return nil, err
		}
		t.domain = platformConfig.DomainName
	}
	return t, nil
}

// Execute 执行 DNS 更新
func (t *DNSTask) Execute(ctx context.Context, recordType string, newValue string) error {
	if t.provider == nil {
		return nil
	}
	t.logger.Info("【DNS】开始更新",
		slog.String("provider", t.provider.Name()),
		slog.String("domain", t.domain),
		slog.String("subdomain", t.subdomain),
		slog.String("recordType", recordType),
		slog.String("newValue", newValue),
		slog.String("config_json", string(t.record_config_json)),
	)

	if len(t.record_config_json) > 0 {
		if err := t.provider.ParseCustomConfig(t.record_config_json); err != nil {
			t.logger.Error("【DNS】解析自定义配置失败",
				slog.String("name", t.provider.Name()),
				slog.Any("err", err),
			)
			return err
		}
	}

	recordID, oldValue, err := t.provider.Check(ctx, t.domain, t.subdomain, recordType)
	if err != nil && !errors.Is(err, ErrRecordNotFound) {
		t.logger.Error("【DNS】检查失败",
			slog.String("recordID", recordID),
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}
	if oldValue == newValue {
		t.logger.Info("【DNS】IP 未改变，无需更新")
		return nil
	}

	if err := t.provider.Update(ctx, recordID, t.domain, t.subdomain, newValue, recordType); err != nil {
		t.logger.Error("【DNS】更新失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	if err := t.provider.Save(ctx, recordID, t.domain, t.subdomain, newValue, recordType); err != nil {
		t.logger.Error("【DNS】保存失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
		return err
	}

	t.logger.Info("【DNS】更新成功")
	return nil
}

// Save 保存 DNS 记录
func (t *DNSTask) Save(ctx context.Context, db *ent.Client) (err error) {
	return nil
}

// Notify 执行 DNS 通知
func (t *DNSTask) Notify(ctx context.Context, id int64, recordType string, webhookService *webhookService.Send) (err error) {
	if webhookService == nil {
		return nil
	}

	// 通知模板数据
	notifyData := utils.TemplateData{
		EventType:    "DNS_UPDATE",
		EventTime:    time.Now().Format("2006-01-02 15:04:05"),
		EventBody:    "DNS记录更新事件",
		RecordDomain: fmt.Sprintf("%s.%s", t.subdomain, t.domain),
		RecordType:   recordType,
	}

	// 发送通知
	if err := webhookService.Send(
		ctx,
		webhooksubscription.ObjTypeDNSRecord,
		id,
		notifyData,
	); err != nil {
		t.logger.Error("【DNS】通知失败",
			slog.String("name", t.provider.Name()),
			slog.Any("err", err),
		)
	}

	return nil
}
